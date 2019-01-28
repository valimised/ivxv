/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Event log browsing page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // fill log table with data
  $('#dataTables-log').DataTable({
      'ajax': '/ivxv/cgi/eventlog',
      stateSave: true,
      'columns': [{
        'data': 'timestamp'
      }, {
        'data': 'service'
      }, {
        'data': 'level'
      }, {
        'data': 'event'
      }, {
        'data': 'message'
      }],
      'order': [
        [0, 'desc']
      ],
      'fnRowCallback': function(nRow, aData) {
        if (aData.level === 'INFO') {
          $(nRow).addClass('success');
        } else {
          $(nRow).addClass('danger');
        }
      }
    })
    .on('xhr', function(e, settings, json) {
      if (json === null) {
        $('#loadstatus')
          .removeClass('text-info')
          .addClass('text-danger')
          .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
      } else {
        $('#loadstatus')
          .removeClass('text-danger')
          .addClass('text-info')
          .html('Andmete laadimise aeg: ' + formatTime(loadDate, 0));
      }
    });
};
