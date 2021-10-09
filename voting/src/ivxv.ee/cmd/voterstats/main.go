/*
The voterstats application is used for exporting voter statistics from the
storage service.
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"ivxv.ee/command"
	"ivxv.ee/command/exit"
	"ivxv.ee/command/status"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	"ivxv.ee/storage"
	//ivxv:modules container
	//ivxv:modules storage
)

const usage = `voterstats exports voter statistics from the collector's storage service for
use by election management services.

voterstats can export two types of voter statistics: a simple running total of
voters (the default) and a detailed per county count of voters per day.

The statistics only contain voters that have cast a complete vote. This might
be less than the number of votes exported using voteexp, because the latter
exports some partial votes needed by the processing application.`

var (
	detailedp = flag.Bool("detailed", false, "export detailed statistics")

	timezonep = flag.String("timezone", "Local",
		"name of IANA Time Zone location to use for partitioning votes by day,\n"+
			"e.g., Europe/Tallinn")

	qp = flag.Bool("q", false, "quiet, do not show progress")

	progress *status.Line
)

func main() {
	// Call voterstatsmain in a separate function so that it can set up
	// defers and have them trigger before returning with a non-zero exit
	// code.
	os.Exit(voterstatsmain())
}

const (
	countyForeign = "FOREIGN"
	countyTotal   = "TOTAL"
)

func voterstatsmain() (code int) {
	c := command.New("ivxv-voterstats", usage, "output file")
	defer func() {
		code = c.Cleanup(code)
	}()

	if !*qp {
		progress = status.New()
	}

	// Parse start and stop time for statistics period calculation.
	var start, stop time.Time
	var err error
	if start, err = c.Conf.Election.ElectionStartTime(); err != nil {
		return c.Error(exit.Config, ElectionStartTimeError{Err: err},
			"failed to parse election start time:", err)
	}
	if stop, err = c.Conf.Election.ServiceStopTime(); err != nil {
		return c.Error(exit.Config, ServiceStopTimeError{Err: err},
			"failed to parse service stop time:", err)
	}
	if start.After(stop) {
		return c.Error(exit.Config,
			ElectionStartTimeAfterServiceStopError{Start: start, Stop: stop},
			"election start time after service stop:", start, ">", stop)
	}

	if c.Until < command.CheckInput {
		return exit.OK
	}

	// Load time zone data for the location.
	loc, err := time.LoadLocation(*timezonep)
	if err != nil {
		return c.Error(exit.DataErr, LoadTimeZoneError{Err: err},
			"failed to load time zone:", err)
	}

	// Load counties list from storage and parse into dists.
	dists := new(districtlist)
	counties, err := c.Storage.GetCounties(c.Ctx)
	if err != nil {
		code = exit.Software
		if errors.CausedBy(err, new(storage.NotExistError)) != nil {
			code = exit.NoInput
		}
		return c.Error(code, GetCountiesError{Err: err},
			"failed to get counties list from storage:", err)
	}
	if err := json.Unmarshal(counties, &dists.Counties); err != nil {
		return c.Error(exit.DataErr, ParseCountiesError{Err: err},
			"failed to parse counties list:", err)
	}
	if _, ok := dists.Counties[countyForeign]; ok {
		return c.Error(exit.DataErr, ForeignCountyError{},
			"counties list contains county named", countyForeign)
	}
	if _, ok := dists.Counties[countyTotal]; ok {
		return c.Error(exit.DataErr, TotalCountyError{},
			"counties list contains county named", countyTotal)
	}

	if c.Until < command.Execute {
		return exit.OK
	}

	// Create the output file before exporting. Avoid defaulting to
	// standard output, because progress is printed there.
	output := c.Args[0]
	fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return c.Error(exit.CantCreate, CreateOutputError{Output: output, Err: err},
			"failed to create output file:", err)
	}
	defer func() {
		if err := fp.Close(); err != nil && code == exit.OK {
			code = c.Error(exit.IOErr, CloseOutputError{Err: err},
				"failed to close output file:", err)
		}
	}()

	var stats interface{}
	if *detailedp {
		periods := cumulativePeriods(start, stop, loc)
		stats, err = statsDetailed(c.Ctx, c.Conf.Election.Identifier,
			dists, periods, c.Storage)
	} else {
		stats, err = statsTotal(c.Ctx, c.Conf.Election.Identifier, c.Storage)
	}
	if err != nil {
		return c.Error(exit.Unavailable, ExportStatisticsError{Err: err},
			"failed to export statistics:", err)
	}

	// Write exported statistics as JSON to output.
	encoder := json.NewEncoder(fp)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(stats)
	if err != nil {
		return c.Error(exit.Unavailable, EncodeStatisticsError{Err: err},
			"failed to encode voter statistics:", err)
	}

	return
}

type votersTotal struct {
	Total votersTotalInner `json:"TOTAL"`
}

type votersTotalInner struct {
	Time     string `json:"time"`
	Election string `json:"election"`
	Voted    uint64 `json:"online-voters"`
}

func statsTotal(ctx context.Context, election string, s *storage.Client) (*votersTotal, error) {
	log.Log(ctx, ExportingTotalStatistics{})
	progress.Static("Exporting total statistics:")
	addprogress := progress.Count(0, true)
	progress.Redraw()
	defer progress.Keep()

	var voted uint64
	statsc, errc := s.GetVotedStats(ctx)
	for range statsc {
		voted++
		addprogress(1)
	}
	if err := <-errc; err != nil {
		return nil, GetTotalVotedStatsError{Err: err}
	}
	return &votersTotal{
		Total: votersTotalInner{
			Time:     time.Now().Format(time.RFC3339),
			Election: election,
			Voted:    voted,
		},
	}, nil
}

type votersDetailed struct {
	Election string                 `json:"election"`
	Time     string                 `json:"time"`
	Counties []votersDetailedCounty `json:"data"`
}

type votersDetailedCounty struct {
	County  string                 `json:"county"`
	Periods []votersDetailedPeriod `json:"online-voters"`
}

type votersDetailedPeriod struct {
	Label string `json:"day-order"`
	Voted uint64 `json:"voted-count"`
}

func statsDetailed(ctx context.Context, election string,
	dists *districtlist, periods []period, s *storage.Client) (
	*votersDetailed, error) {

	log.Log(ctx, ExportingDetailedStatistics{})
	progress.Static("Exporting detailed statistics:")
	addprogress := progress.Count(0, true)
	progress.Redraw()
	defer progress.Keep()

	// Map from county to period label to voter count for fast aggregation.
	counts := make(map[string]map[string]uint64, len(dists.Counties)+2)
	addCount := func(county string, labels []string) {
		periodMap, ok := counts[county]
		if !ok {
			periodMap = make(map[string]uint64, len(periods))
			counts[county] = periodMap
		}
		for _, label := range labels {
			periodMap[label]++
		}
	}

	// Get successful voter stats from storage and add to map.
	statsc, errc := s.GetVotedStats(ctx)
	for stats := range statsc {
		labels := findPeriodLabels(periods, stats.Time)
		if len(labels) == 0 {
			nonfatal(ctx, TimeOutsideStatsPeriodError{
				Voter: stats.Voter,
				Time:  stats.Time,
			}, "submission time outside of statistics period:", stats.Time)
			continue
		}
		addCount(countyTotal, labels)
		addprogress(1)

		if stats.AdminCode != "" {
			county := dists.findCounty(stats.AdminCode)
			if county == "" {
				nonfatal(ctx, AdminUnitWithoutCountyError{
					Voter:     stats.Voter,
					AdminCode: stats.AdminCode,
				}, "administrative unit without county:", stats.AdminCode)
				continue
			}
			addCount(county, labels)
		}
	}
	if err := <-errc; err != nil {
		return nil, GetDetailedVotedStatsError{Err: err}
	}

	// Convert from map to desired data structure.
	convertCount := func(county string) votersDetailedCounty {
		periodMap, ok := counts[county]

		// Range over periods slice instead of map to preserve ordering.
		converted := make([]votersDetailedPeriod, len(periods))
		for i, period := range periods {
			count := votersDetailedPeriod{Label: period.label}
			if ok {
				count.Voted = periodMap[period.label]
			}
			converted[i] = count
		}
		return votersDetailedCounty{
			County:  county,
			Periods: converted,
		}
	}

	// The order of counties is lost since they were unmarshaled into a
	// map, but at least sort them to ensure consistent output.
	counties := make([]votersDetailedCounty, 0, len(dists.Counties)+2)
	for _, county := range dists.sortedCounties() {
		counties = append(counties, convertCount(county))
	}
	counties = append(counties, convertCount(countyForeign))
	counties = append(counties, convertCount(countyTotal))

	return &votersDetailed{
		Election: election,
		Time:     time.Now().Format(time.RFC3339),
		Counties: counties,
	}, nil
}

// districtlist is a subset of the election districtlist list. It only contains
// the counties subsection necessary for separating votes by county.
type districtlist struct {
	// Counties maps from county code to voter administrative unit codes
	// that belong to that county.
	Counties map[string][]string

	// countyMap maps from voter administrative unit codes to the county
	// they belong to. Used to speed up findCounty.
	countyMap map[string]string
}

func (d *districtlist) findCounty(unitCode string) string {
	if unitCode == countyForeign {
		return unitCode
	}
	if d.countyMap == nil {
		d.countyMap = make(map[string]string)
		for county, codes := range d.Counties {
			for _, code := range codes {
				d.countyMap[code] = county
			}
		}
	}
	return d.countyMap[unitCode]
}

func (d *districtlist) sortedCounties() []string {
	counties := make([]string, 0, len(d.Counties))
	for county := range d.Counties {
		counties = append(counties, county)
	}
	sort.Strings(counties)
	return counties
}

func nonfatal(ctx context.Context, err log.ErrorEntry, a ...interface{}) {
	log.Error(ctx, NonFatalError{Err: err})
	progress.Hide()
	fmt.Fprintln(os.Stderr, append([]interface{}{"error: non-fatal error:"}, a...)...)
	progress.Show()
}
