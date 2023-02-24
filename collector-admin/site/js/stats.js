/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Stats page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // manage districts selection
  if ($('#districts option').length == 1) {
    $.getJSON('data/districts.json', function(data) {
        /*
         * HTTP GET on https://admin.?.ivxv.ee/ivxv/data/districts.json
         * HTTP GET response that contains HTML tags is not allowed!
         * data is always an JSON object
         */
        data = sanitizeJSON(data);
        var dropdown = $('#districts');
        $.each(data, function() {
          dropdown.append($('<option />').val(this[0]).text(this[1]));
        })
      })
      .fail(function() {
        $('#districts option').text('Ringkondade nimekiri pole saadaval');
      });
  }

  // fill page with stats
  $.getJSON('data/stats.json', function(data) {
      /*
       * HTTP GET on https://admin.?.ivxv.ee/ivxv/data/stats.json
       * HTTP GET response that contains HTML tags is not allowed!
       * data is always an JSON object
       */
      data = sanitizeJSON(data)
      $('#stats-error').hide();
      $('#stats-error-msg').html();
      $('#auth-os').empty();
      $('#table-revoters').empty();
      $('#table-countries').empty();
      $.each(data, function(key, val) {
        if (key === 'data') {
          var district_stats = $(val).prop($('#districts').val());
          $.each(district_stats, function(stats_key, stats_val) {
            if (typeof(stats_val) === 'object') {
              if (stats_key === 'authentication-methods') {
                $('#auth-os').append(
                  '<tr class="info">\n' +
                  '<th colspan="2">Autentimisvahend</th>\n' +
                  '</tr>\n'
                )
              } else if (stats_key === 'operating-systems') {
                $('#auth-os').append(
                  '<tr class="info">\n' +
                  '<th colspan="2">Operatsioonisüsteem</th>\n' +
                  '</tr>\n'
                )
              } else if (stats_key === 'top-10-revoters') {
                $('#table-revoters').append(
                  '<tr class="info">\n' +
                  '<th colspan="2">TOP 10 korduvhääletajat</th>\n' +
                  '</tr>'
                )
              }
              $.each(stats_val, function(stats_table_key, stats_table_val) {
                if (stats_key === 'authentication-methods') {
                  var method = 'ID-kaart';
                  if (stats_table_val[0] === 'ticket') {
                    method = 'Mobiil-ID/Smart-ID';
                  }
                  $('#auth-os').append(
                    '<tr><td>' +
                    method +
                    '</td><td>' +
                    sanitizePrimitive(stats_table_val[1]) +
                    '</td></tr>'
                  )
                } else if (stats_key === 'operating-systems') {
                  $('#auth-os').append(
                    '<tr><td>' +
                    sanitizePrimitive(stats_table_val[0]) +
                    '</td><td>' +
                    sanitizePrimitive(stats_table_val[1]) +
                    '</td></tr>'
                  )
                } else if (stats_key === 'top-10-revoters') {
                  $('#table-revoters').append(
                    '<tr><td>' +
                    sanitizePrimitive(stats_table_val[0]) +
                    '</td><td>' +
                    sanitizePrimitive(stats_table_val[1]) +
                    '</td></tr>'
                  )
                } else if (stats_key === 'votes-by-country') {
                  $('#table-countries').append(
                    '<tr><td>' +
                    sanitizePrimitive(stats_table_val[0]) +
                    '</td><td>' +
                    sanitizePrimitive(stats_table_val[1]) +
                    '</td></tr>'
                  )
                }
              });
            } else {
              $('#' + sanitizePrimitive(stats_key)).text(stats_val);
            }
          });
        } else if (key === 'error') {
          $('#stats-error').show();
          $('#stats-error-msg').html(data[key].replace(/\n/g, '<br />'));
        }
      });

      // data loading stats
      var genDate = new Date();
      genDate.setTime(Date.parse(data['meta']['time_generated']));
      $('#loadstatus')
        .removeClass('text-danger')
        .addClass('text-info')
        .html('Andmete laadimise aeg: ' + formatTime(loadDate, 0) + '<br />' +
          'Andmete genereerimise aeg: ' +
          genDate.toLocaleTimeString(
            'et-EE', {
              year: 'numeric',
              month: 'numeric',
              day: 'numeric'
            }));
    })
    .fail(function() {
      $('#loadstatus')
        .removeClass('text-info')
        .addClass('text-danger')
        .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
      showErrorMessage('Viga seisundi laadimisel', true);
    });
}
