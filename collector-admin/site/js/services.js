/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Services management page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // data mappings
  var service_states = {
    'NOT INSTALLED': ['Paigaldamata', 'warning', 0],
    'INSTALLED': ['Paigaldatud', '', 0],
    'CONFIGURED': ['Seadistatud', 'success', 0],
    'FAILURE': ['Tõrge', 'danger', 0],
    'REMOVED': ['Eemaldatud', '', 0]
  };
  var service_types = {
    'backup': 'Varundusteenus',
    'choices': 'Nimekirjateenus',
    'mid': 'Mobiil-ID abiteenus',
    'smartid': 'Smart-ID abiteenus',
    'proxy': 'Vahendusteenus',
    'storage': 'Talletusteenus',
    'log': 'Logikogumisteenus',
    'votesorder': 'Järjekorrateenus',
    'voting': 'Hääletamisteenus',
    'verification': 'Kontrollteenus'
  };

  // load collector state
  $.getJSON('data/status.json', function(state) {
    /*
     * HTTP GET on https://admin.?.ivxv.ee/ivxv/data/status.json
     * HTTP GET response that contains HTML tags is not allowed!
     * state is always an JSON object
     */
    state = sanitizeJSON(state);
    hideErrorMessage();

    var i = 1;
    $('#service_list_table').toggle(state['service'] !== undefined);

    // register service details block visibility states
    var is_service_visible = {};
    $.each(state['service'], function(k, v) {
      var details_block_id = k.replace(/[@\.]/g, '_') + '_details';
      is_service_visible[details_block_id] = $('#' + details_block_id).is(':visible');
    });

    $('#services-list').empty();
    $.each(state['service'], function(k, v) {
      var zone = 'Tundmatu';
      var state_str = 'Tundmatu';
      var status_class = 'warning';
      var service_type = v['service-type'] in service_types ? service_types[v['service-type']] : 'Tundmatu';
      var show_mobile = 'mid-token-key' in state['service'][k] ? '' : 'display: none';
      var show_tlskey = 'tls-key' in state['service'][k] ? '' : 'display: none';
      var show_tlscert = 'tls-cert' in state['service'][k] ? '' : 'display: none';
      var lastdata = state['service'][k]['last-data'] ? state['service'][k]['last-data'] : '-';
      var econf = state['service'][k]['election-conf-version'] ? state['service'][k]['election-conf-version'] : '-';
      var tlsc = state['service'][k]['tls-cert'] ? state['service'][k]['tls-cert'] : '-';
      var tlsk = state['service'][k]['tls-key'] ? state['service'][k]['tls-key'] : '-';
      var midk = state['service'][k]['mid-token-key'] ? state['service'][k]['mid-token-key'] : '-';
      var pingerr = state['service'][k]['ping-errors'] ? state['service'][k]['ping-errors'] : '-';
      var bg_info = state['service'][k]['bg_info'];
      var tconf = state['service'][k]['technical-conf-version'] ? state['service'][k]['technical-conf-version'] : '-';
      var ipaddr = state['service'][k]['ip-address'] ? state['service'][k]['ip-address'] : '-';

      // Service field in json doesn't have info about its zone, have to find it by going through all of them
      $.each(state['network'], function(network_key, network_value) {
        if (k in network_value) {
          zone = network_key;
          return false;
        }
      });

      if (v['state'] in service_states) {
        service_states[v['state']][2]++;
        state_str = service_states[v['state']][0];
        status_class = service_states[v['state']][1];
      }

      var details_block_id = k.replace(/[@\.]/g, '_') + '_details';
      var details_block_style = is_service_visible[details_block_id] ? '' : 'display: none;';
      $('#services-list').append(
        '<tr class="' + status_class + '" onclick="$(this).next().toggle()">' +
        '<td>' + i + '</td>' +
        '<td>' + k + '</td>' +
        '<td>' + zone + '</td>' +
        '<td>' + service_type + '</td>' +
        '<td>' + state_str + '</td>' +
        '</tr>' +
        '<tr class="warning" ' +
        'style="' + details_block_style + '" ' +
        'id="' + details_block_id + '" ' +
        '>' +
        '<td colspan="5">' +
        '<table class="table table-striped">' +
        '<tbody>' +
        '<tr>' +
        '<td align="right">Vigade arv:</td>' +
        '<td>' + pingerr + '</td>' +
        '</tr>' +
        '<tr>' +
        '<td align="right">Viimane kontroll:</td>' +
        '<td>' + lastdata + '</td>' +
        '</tr>' +
        '<tr>' +
        '<td align="right">Tehniline seadistus:</td>' +
        '<td>' + tconf + '</td>' +
        '</tr>' +
        '<tr>' +
        '<td align="right">Valimiste seadistus:</td>' +
        '<td>' + econf + '</td>' +
        '</tr>' +
        '<tr>' +
        '<td align="right">IP-aaddress:</td>' +
        '<td>' + ipaddr + '</td>' +
        '</tr>' +
        '<tr style="' + show_tlscert + ';">' +
        '<td align="right">TLS-sertifikaat:</td>' +
        '<td>' + tlsc + '</td>' +
        '</tr>' +
        '<tr style="' + show_tlskey + ';">' +
        '<td align="right">TLS-võti:</td>' +
        '<td>' + tlsk + '</td>' +
        '</tr>' +
        '<tr style="' + show_mobile + ';">' +
        '<td align="right">Mobiil-ID krüptimissaladus:</td>' +
        '<td>' + midk + '</td>' +
        '</tr>' +
        '</tbody>' +
        '</table>' +
        '</td>' +
        '</tr>'
      );
      if (bg_info) {
        $('#services-list').append(
          '<tr>' +
          '<td/>' +
          '<td colspan="4" class="warning">' +
          '  <small>Taustainfo: </small>' +
          '  <small class="text-info">' + bg_info.replace(/\n/g, '<br />') + '</small>' +
          '</td>' +
          '</tr>'
        );
      }
      i++;
    });

    // data loading stats
    var genDate = new Date();
    genDate.setTime(Date.parse(state['meta']['time_generated']));
    $('#loadstatus')
      .removeClass('text-danger')
      .addClass('text-info')
      .html('Andmete laadimise aeg: ' + formatTime(loadDate, 0) + '<br />' +
        'Andmete genereerimise aeg: ' + genDate.toLocaleTimeString('et-EE', {}));

  }).done(function() {
    $('#not_installed').html(service_states['NOT INSTALLED'][2]);
    $('#configured').html(service_states['CONFIGURED'][2]);
    $('#installed').html(service_states['INSTALLED'][2]);
    $('#failure').html(service_states['FAILURE'][2]);
    $('#removed').html(service_states['REMOVED'][2]);
  }).fail(function() {
    $('#loadstatus')
      .removeClass('text-info')
      .addClass('text-danger')
      .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
    showErrorMessage('Viga seisundi laadimisel', true);
  });
}
