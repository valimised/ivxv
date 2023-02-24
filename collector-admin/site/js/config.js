/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Configuration status page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // load collector state
  $.getJSON('data/status.json', function(state) {
      /*
       * HTTP GET on https://admin.?.ivxv.ee/ivxv/data/status.json
       * HTTP GET response that contains HTML tags is not allowed!
       * state is always an JSON object
       */
      state = sanitizeJSON(state);
      display_cfg_panel(
        'trust', state, state['config']['trust'], 'Usaldusjuure seadistus');
      display_cfg_panel(
        'technical', state, state['config-apply']['technical'], 'Tehniline seadistus');
      display_cfg_panel(
        'election', state, state['config-apply']['election'], 'Valimiste seadistus');
      display_cfg_panel(
        'choices', state, state['config-apply']['choices'], 'Valikute nimekiri');
      display_cfg_panel(
        'districts', state, state['list']['districts'], 'Ringkondade nimekiri');
      var changeset_no;
      display_cfg_panel(
        'voters0000', state, state['config-apply']['voters0000'], 'Valijate nimekiri (algne)');
      for (var i = 1; i < 10000; i++) {
        var iStr = 'voters' + String(i).padStart(4, '0');
        if (!(iStr + '-state' in state['list']))
          break;
        if (iStr in state['config-apply']) {
          display_cfg_panel(
            iStr, state, state['config-apply'][iStr],
            'Valijate muudatusnimekiri nr. ' + i);
        } else {
          display_cfg_panel(
            iStr, state, state['list'][iStr],
            'Valijate muudatusnimekiri nr. ' + i);
        }
      }

      hideErrorMessage();

      // data loading stats
      var genDate = new Date();
      genDate.setTime(Date.parse(state['meta']['time_generated']));
      $('#loadstatus')
        .removeClass('text-danger')
        .addClass('text-info')
        .html(
          'Andmete laadimise aeg: ' + formatTime(loadDate, 0) + '<br />' +
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

var state_filenames = {};

/**
 * Display config state panel
 *
 * @param {string} id_prefix
 * @param {Object} state
 * @param {Object} cfg
 * @param {string} title
 */
function display_cfg_panel(id_prefix, state, cfg, title) {
  id_prefix = sanitizePrimitive(id_prefix);
  state = sanitizeJSON(state);
  cfg = sanitizeJSON(cfg);
  // Create panel if required
  var panel = $('#' + id_prefix + '-cfg-state-panel');
  if (!panel.length) {
    $('#upload-row').before(
      '<div class="row">' +
      '    <div class="col-lg-12">' +
      '        <div id="' + id_prefix + '-cfg-state-panel" class="panel">' +
      '            <div class="panel-heading">' +
      '                <h4 class="panel-title">' + sanitizePrimitive(title) + '</h4>' +
      '            </div>' +
      '            <div class="panel-body">' +
      '              <div />' + // Placeholder for config info text
      '              <button type="button" class="btn btn-default" onClick="toggle_apply_log(this);">Rakendamise logi</button>' +
      '              <pre style="display: none;" class="pre-scrollable" />' + // Placeholder for log content
      '            </div>' +
      '        </div>' +
      '    </div>' +
      '</div>'
    );
    panel = $('#' + id_prefix + '-cfg-state-panel');
  }

  // Configure panel
  var panel_body = panel.find('.panel-body');
  panel
    .removeClass('panel-green')
    .removeClass('panel-warning')
    .removeClass('panel-danger');
  if ((id_prefix === 'trust') || (id_prefix === 'districts')) {
    var ver_element_id = 'cfg-ver-' + id_prefix;
    panel.addClass(cfg === null ? 'panel-danger' : 'panel-green');
    panel_body
      .find('div:first')
      .html(
        '<div>Seisund: ' + (cfg === null ? 'Laadimata' : 'Rakendatud haldusteenusele') + '</div>' +
        '<div>' +
        (cfg === null ? '-' : 'Versioon: <span id="' + ver_element_id + '"></span>') +
        '</div>');
    outputCmdVersion('#' + ver_element_id, id_prefix, state)
    panel_body.find('button').hide();
  } else if (cfg === undefined) {
    panel.addClass('panel-danger');
    panel_body.find('div:first').html('<div>Laadimata</div>');
    panel_body.find('button').hide();
  } else {
    state_filenames[id_prefix] = cfg['state_file'];
    if (cfg['completed']) {
      panel.addClass('panel-green');
    } else {
      panel.addClass('panel-warning');
    }

    if (id_prefix.startsWith('voters')) {
      stateStr = voterListStateDescriptions.get(state['list'][id_prefix + '-state']);
    } else {
      stateStr = cfg['completed'] ? 'rakendatud' : 'rakendamisel';
    }
    if (id_prefix.startsWith('voters') && !(id_prefix in state['config-apply'])) {
      panel_body
        .find('div:first')
        .html(
          '<div>Seisund: ' + sanitizePrimitive(stateStr) + '</div>' +
          '<div>Rakendatav versioon: <span id="cfg-ver-' + id_prefix + '">[ määramata ]</a></div>' +
          '<div>Rakendamise katseid: 0</div>'
        );
    } else {
      panel_body
        .find('div:first')
        .html(
          '<div>Seisund: ' + sanitizePrimitive(stateStr) + '</div>' +
          '<div>Rakendatav versioon: <span id="cfg-ver-' + id_prefix + '">' + cfg['version'] + '</a></div>' +
          '<div>Rakendamise katseid: ' + cfg['attempts'] + '</div>'
        );
      outputCmdVersion('#cfg-ver-' + id_prefix, id_prefix, state);
    }

    if (cfg['attempts']) {
      panel_body.find('button').show('slow');
      var logbox = panel_body.find('pre');
      if (logbox.is(':visible')) {
        refresh_log(logbox, cfg['state_file']);
      }
    } else {
      panel_body.find('button').hide('slow');
      panel_body.find('pre').hide('slow');
    }
  }
}

/**
 * Toggle config log box
 */
function toggle_apply_log(toggle_button) {
  var parent_element = $(toggle_button).parent();
  var logbox = parent_element.find('pre');
  logbox.toggle();
  if (logbox.is(':visible')) {
    var cfg_type = parent_element.parent().attr('id').replace('-cfg-state-panel', '');
    refresh_log(logbox, state_filenames[cfg_type]);
  }
}


/**
 * Refresh logbox content
 */
function refresh_log(logbox, filename) {
  var url = '/ivxv/data/commands/' + filename;
  $.getJSON(encodeURI(url), function(state) {
      /*
       * HTTP GET on https://admin.?.ivxv.ee/ivxv/data/commands/??
       * HTTP GET response that contains HTML tags is not allowed!
       * state is always an JSON object
       */
      state = sanitizeJSON(state);
      logbox.text(state['log'][state['attempts'] - 1].join('\n'));
    })
    .fail(function(response) {
      logbox.text(sanitizePrimitive(response.responseText));
    });
}

/**
 * Reset upload form
 */
function reset_upload_form() {
  $('input[type=file]').val(null);
  $('#file-upload-submit').attr('disabled', '');
}

// Variable to store uploaded files
var files;

/**
 * Grab the files and set them to our variable
 */
function prepareUpload(event) {
  files = event.target.files;
  $('#file-upload-submit').attr('disabled', null);
  $('#upload-message').hide();
}

/**
 * Catch the form submit and upload the files
 */
function uploadFiles(event) {
  $('#upload-message')
    .removeClass('alert-danger')
    .removeClass('alert-success')
    .hide();

  event.stopPropagation(); // Stop stuff happening
  event.preventDefault(); // Totally stop stuff happening
  // Create a formdata object and add the files
  var data = new FormData();
  data.append('upload', files[0]);
  data.append('type', sanitizePrimitive($('#drop').find(':selected').val()));

  var form = $('#config-upload-form');
  $.ajax({
    url: encodeURI(form.attr('action')),
    type: sanitizePrimitive(form.attr('method')),
    data: data,
    cache: false,
    dataType: 'json',
    processData: false, // Don't process the files
    contentType: false, // Set content type to false as jQuery will tell the server its a query string request

    // Success
    success: function(data, textStatus, jqXHR) {
      console.log(jqXHR.responseJSON.message);
      $('#upload-message')
        .html(
          sanitizePrimitive(jqXHR.responseJSON.message) +
          '<hr />' +
          '<pre>' + sanitizePrimitive(jqXHR.responseJSON.log.join('\n')) + '</pre>'
        )
        .addClass(jqXHR.responseJSON.success ? 'alert-success' : 'alert-danger')
        .show();
      reset_upload_form();
    },

    // Handle errors
    error: function(jqXHR, textStatus, errorThrown) {
      console.log(jqXHR);
      $('#upload-message')
        .html(sanitizePrimitive(jqXHR.responseText))
        .addClass('alert-danger')
        .show();
    }
  });
}
