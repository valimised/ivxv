/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Overview page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // data mappings
  var states = {
    'NOT INSTALLED': ['Paigaldamata', 'warning', 0],
    'INSTALLED': ['Paigaldatud', '', 0],
    'CONFIGURED': ['Seadistatud', 'success', 0],
    'PARTIAL FAILURE': ['Osaline tõrge', 'danger', 0],
    'FAILURE': ['Tõrge', 'danger', 0],
    'REMOVED': ['Eemaldatud', '', 0]
  };
  var phases = {
    'PREPARING': 'Ettevalmistamine',
    'WAITING FOR SERVICE START': 'Teenuse käivitamise ootamine',
    'WAITING FOR ELECTION START': 'Valimise alguse ootamine',
    'ELECTION': 'Häälte kogumine',
    'WAITING FOR SERVICE STOP': 'Teenuse seiskamise ootamine',
    'FINISHED': 'Lõpetatud'
  };

  // load collector state
  $.getJSON('data/status.json', function(state) {
      hideErrorMessage();

      // election ID
      $('#electionid').toggle(state['election']['election-id'] !== null);
      $('#election-id').text(state['election']['election-id']);

      // collector state
      $('#collector-state')
        .html(
          state['collector']['state'] in states ?
          states[state['collector']['state']][0] :
          'Tundmatu');
      $('#collector-state-panel')
        .removeClass('panel-danger')
        .removeClass('panel-warning')
        .removeClass('panel-success')
        .addClass(
          state['collector']['state'] in states ?
          'panel-' + states[state['collector']['state']][1] :
          'panel-danger');

      // voting phase
      $('#voting-phase').toggle(state['config']['election'] !== null);
      $('#voting-stage')
        .text(
          state['election']['phase'] in phases ?
          phases[state['election']['phase']] :
          'Tundmatu');
      $('#voting-stage-start').text(state['election']['phase-start']);
      $('#voting-stage-end').text(state['election']['phase-end']);

      // config packages
      $('#trust')
        .addClass('list-group-item-warning')
        .removeClass('list-group-item-success');
      if (state['config']['trust']) {
        outputCmdVersion('#config-trust', 'trust', state)
        $('#trust')
          .removeClass('list-group-item-warning')
          .addClass('list-group-item-success');
      }
      $('#tech')
        .addClass('list-group-item-warning')
        .removeClass('list-group-item-success');
      if (state['config']['technical']) {
        outputCmdVersion('#config-tech', 'technical', state)
        $('#tech')
          .removeClass('list-group-item-warning')
          .addClass('list-group-item-success');
      }
      $('#election')
        .addClass('list-group-item-warning')
        .removeClass('list-group-item-success');
      if (state['config']['election']) {
        outputCmdVersion('#config-election', 'election', state)
        $('#election')
          .removeClass('list-group-item-warning')
          .addClass('list-group-item-success');
      }

      // voting lists - choices
      outputCmdVersion('#list-choices', 'choices', state)
      $('#choiceslist')
        .removeClass('list-group-item-danger')
        .removeClass('list-group-item-success')
        .removeClass('list-group-item-warning');
      $('#choicesliststatus')
        .removeClass('list-group-item-danger')
        .removeClass('list-group-item-success')
        .removeClass('list-group-item-warning');
      if (!state['list']['choices']) {
        $('#list-choices-status').text('Laadimata');
        $('#choiceslist').addClass('list-group-item-danger');
        $('#choicesliststatus').addClass('list-group-item-danger');
      } else if (!state['list']['choices-loaded']) {
        $('#list-choices-status').text('Laaditud haldusteenusesse');
        $('#choiceslist').addClass('list-group-item-warning');
        $('#choicesliststatus').addClass('list-group-item-warning');
      } else if (state['list']['choices'] === state['list']['choices-loaded']) {
        $('#list-choices-status').text('Rakendatud kogumisteenusele');
        $('#choiceslist').addClass('list-group-item-success');
        $('#choicesliststatus').addClass('list-group-item-success');
      }

      // voting lists - districts
      outputCmdVersion('#list-districts', 'districts', state)
      $('#districtslist')
        .removeClass('list-group-item-danger')
        .removeClass('list-group-item-success');
      $('#districtsliststatus')
        .removeClass('list-group-item-danger')
        .removeClass('list-group-item-success');
      if (state['list']['districts']) {
        $('#districtslist').addClass('list-group-item-success');
        $('#districtsliststatus').addClass('list-group-item-success');
        $('#list-districts-status').text('Laaditud haldusteenusesse');
      } else {
        $('#districtslist').addClass('list-group-item-danger');
        $('#districtsliststatus').addClass('list-group-item-danger');
        $('#list-districts-status').text('-');
      }

      // voting lists - voters
      fillVoterListStateCounters(state['list']);
      $('#voterslist')
        .removeClass('list-group-item-danger')
        .removeClass('list-group-item-success');
      if ((state['list']['voters-list-total'] === 0) ||
        (state['list']['voters-list-invalid'] !== 0)) {
        $('#voterslist').addClass('list-group-item-danger');
      } else if (state['list']['voters-list-pending'] === 0) {
        $('#voterslist').addClass('list-group-item-success');

        $('#list-list').empty();
        for (var changeset_no = 0; changeset_no < 10000; changeset_no++) {
          var iStr = 'voters' + String(changeset_no).padStart(4, '0');
          if (!(iStr + '-state' in state['list']))
            break;
          var listStatus = voterListStateDescriptions.get(state['list'][iStr + '-state']);
          $('#list-list').append(
            '<li class="list-group-item list-group-item-success" style="padding-left:25px">' +
            (changeset_no + 1) + '. ' + listStatus + ': ' + state['list'][iStr] +
            '</li>'
          );
        }
      } else {
        $('#voterslist').addClass('list-group-item-warning');
      }

      // service summary
      var serviceexists = false;
      $.each(state['service'], function(serviceName, service) {
        if (service['state'] in states) {
          states[service['state']][2]++;
          serviceexists = true;
        }
      });
      $('#service_summary').toggle(serviceexists);
      if (serviceexists) {
        $('#not_installed').html(states['NOT INSTALLED'][2]);
        $('#configured').html(states['CONFIGURED'][2]);
        $('#installed').html(states['INSTALLED'][2]);
        $('#failure').html(states['FAILURE'][2]);
        $('#p_failure').html(states['PARTIAL FAILURE'][2]);
        $('#removed').html(states['REMOVED'][2]);
      }

      // debian packages
      $('#debs-exist-count').text(state['storage']['debs_exists'].length);
      if (state['storage']['debs_missing'].length === 0) {
        $('#packagepanel')
          .removeClass('panel-danger')
          .addClass('panel-success')
      } else {
        $('#packagepanel')
          .removeClass('panel-success')
          .addClass('panel-danger')
      }
      $('#debs-missing-count').text(state['storage']['debs_missing'].length);
      $('#missing-debs-list').empty();
      for (var i = 0; i < state['storage']['debs_missing'].length; i++) {
        $('#missing-debs-list').append(
          '<div class="panel panel-danger">' +
          '<div class="panel-heading">' +
          state['storage']['debs_missing'][i] +
          '</div>' +
          '</div>'
        );
      }

      // users
      var usercount = 0;
      if (state['user']) {
        usercount = Object.keys(state['user']).length;
      }
      $('#user-count').text(usercount);

      // command packages
      $('#command-files-count').text(
        state['storage']['command_files_active'].length +
        state['storage']['command_files_inactive'].length
      );
      $('#command-files-active-count').text(state['storage']['command_files_active'].length);
      $('#command-files-inactive-count').text(state['storage']['command_files_inactive'].length);

      // data loading stats
      var genDate = new Date();
      genDate.setTime(Date.parse(state['meta']['time_generated']));
      $('#loadstatus')
        .removeClass('text-danger')
        .addClass('text-info')
        .html('Andmete laadimise aeg: ' + formatTime(loadDate, 0) + '<br />' +
          'Andmete genereerimise aeg: ' + genDate.toLocaleTimeString('et-EE', {}));
    })
    .fail(function() {
      $('#loadstatus')
        .removeClass('text-info')
        .addClass('text-danger')
        .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
      showErrorMessage('Viga seisundi laadimisel', true);
    });
}
