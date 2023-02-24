/*
 * IVXV Internet voting framework
 *
 * Administrator interface - User management page
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
      hideErrorMessage();

      var i = 1;

      $('#user-list').empty();
      $.each(state['user'], function(k, v) {
        $('#user-list').append(
          '<tr>' +
          '<td>' + i + '</td>' +
          '<td>' + k + '</td>' +
          '<td>' + v + '</td>' +
          '</tr>'
        );

        i++;
      });

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
  data.append('type', 'user');

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
      loadPageData()
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
