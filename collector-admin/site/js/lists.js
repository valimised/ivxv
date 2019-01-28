/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Voting lists management page
 */

/**
 * Load page data
 */
function loadPageData() {
  var loadDate = new Date();
  loadDate.setTime(Date.now());

  // load collector state
  $.getJSON('data/status.json', function(state) {
      hideErrorMessage();

      // choices list
      outputCmdVersion('#list-choices', 'choices', state)
      if (!state['list']['choices']) {
        $('#list-choices-status').text('Laadimata');
        $('#panel-choices-list').attr('class', 'panel panel-red');
      } else {
        $('#choicesoption').hide();
        $('#drop').find('option[value="voters"]').prop('selected', true);
        if (!state['list']['choices-loaded']) {
          $('#list-choices-status').text('Laaditud haldusteenusesse');
          $('#panel-choices-list').attr('class', 'panel panel-warning');
        } else if (state['list']['choices'] === state['list']['choices-loaded']) {
          $('#list-choices-status').text('Rakendatud kogumisteenusele');
          $('#panel-choices-list').attr('class', 'panel panel-success');
        }
      }

      // voters lists
      $('#list-voters-loaded').text(state['list']['voters-list-loaded']);
      $('#list-voters-pending').text(state['list']['voters-list-pending']);
      $('#panel-voters-list').attr('class', 'panel panel-warning');
      $('#remove-voters-lists').hide();
      if (state['list']['voters-list-pending'] > 0) {
        $('#remove-voters-lists').show();
        $('#remove-voters-lists').prop('disabled', false);
      }

      if (state['list']['voters-list-loaded'] === 0) {
        $('#panel-voters-list').attr('class', 'panel panel-red');
      } else if (state['list']['voters-list-pending'] === 0) {
        $('#panel-voters-list').attr('class', 'panel panel-success');

        $('#list-list').empty();
        for (var i = 0; i < state['list']['voters-list-loaded']; i++) {
          var iStr = 'voters' + (i < 10 ? '0' : '') + (i + 1);
          var listStatus =
            state['list'][iStr] === state['list'][iStr + '-loaded'] ?
            'LAADITUD' : 'OOTEL';
          $('#list-list').append(
            '<li class="list-group-item" style="padding-left:25px">' +
            (i + 1) + '. ' + listStatus + ': ' + state['list'][iStr] +
            '</li>'
          );
        }
      }

      // districts list
      outputCmdVersion('#list-districts', 'districts', state)
      if (!state['list']['districts']) {
        $('#list-districts-status').text('Laadimata');
        $('#panel-districts-list').attr('class', 'panel panel-red');
      } else {
        $('#list-districts-status').text('Laaditud haldusteenusesse');
        $('#panel-districts-list').attr('class', 'panel panel-success');
      }

      var genDate = new Date();
      genDate.setTime(Date.parse(state['meta']['time_generated']));
      $('#loadstatus')
        .removeClass('text-danger')
        .addClass('text-info')
        .html('Andmete laadimise aeg: ' + formatTime(loadDate, 0) + '<br />' +
          'Andmete genereerimise aeg: ' + formatTime(genDate, 3));
    })
    .fail(function() {
      $('#loadstatus')
        .removeClass('text-info')
        .addClass('text-danger')
        .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
      showErrorMessage('Viga seisundi laadimisel', true);
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
  $('#upload-message').hide()
    .removeClass('alert-danger')
    .removeClass('alert-success');

  event.stopPropagation(); // Stop stuff happening
  event.preventDefault(); // Totally stop stuff happening
  // Create a formdata object and add the files
  var data = new FormData();
  data.append('upload', files[0]);
  data.append('type', $('#drop').find(':selected').val());

  var form = $('#config-upload-form');
  $.ajax({
    url: form.attr('action'),
    type: form.attr('method'),
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
          jqXHR.responseJSON.message +
          '<hr />' +
          '<pre>' + jqXHR.responseJSON.log.join('\n') + '</pre>'
        )
        .addClass(jqXHR.responseJSON.success ? 'alert-success' : 'alert-danger')
        .show();
      reset_upload_form();
    },
    // Handle errors
    error: function(jqXHR, textStatus, errorThrown) {
      console.log(jqXHR);
      $('#upload-message')
        .html(jqXHR.responseText)
        .addClass('alert-danger')
        .show();
    }
  });
}

/**
 * Send a request to remove loaded but not applied voter lists
 */
function removeVotersLists() {
  $('#remove-voters-lists').prop('disabled', true);
  $.ajax({
    method: 'POST',
    url: '/ivxv/cgi/remove-voters-lists',
    success: function(resp) {
      alert(resp.message);
      loadPageData();
    },
    error: function(resp) {
      alert(resp.message);
      loadPageData();
    }
  });
}
