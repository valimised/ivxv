/*
 * IVXV Internet voting framework
 *
 * Administrator interface - JavaScript helpers
 */

var pageContext = null; // Page context data
var userContext = null; // User context data (subset of pageContext)

/**
 * Query page context data and write it
 * to pageContext and userContext variables.
 */
function getContextData() {
  // query page context data
  console.debug('Loading context data');
  $.getJSON('cgi/context.json', function(context) {
    pageContext = context.data;
    userContext = pageContext['current-user'];
    console.debug('Current user: ' + userContext['cn'] +
                  ' ' + userContext['idcode']);
    console.debug('Current role: ' + userContext['role'] +
                  ' (' + userContext['role-description'] + ')');
    copyObjectToHtml(userContext, 'user-');

    // append voting ID to page title
    if (pageContext['voting']['id']) {
      $('title').prepend(pageContext['voting']['id'] + ' – ');
    }

    // check user permissons and take action
    if (userContext['permissions'].length === 0) {
      // user has no permissons, generate error message and paint navbar to red
      $('.navbar').addClass('label-danger').find('li').addClass('label-danger');
      showErrorMessage('Puuduvad kasutajaõigused');
    }
    else if ("function" === typeof(loadPageData)) {
      // execute loadPageData() if such function exists
      loadPageData();
    }

    // Hide menu items if user has no permissions to access it
    var permission_page_map = [["election-conf-admin", "lists.html"],
                               ["stats-view", "stats.html"],
                               ["user-admin", "users.html"],
                               ["election-conf-admin", "tech-conf-admin", "services.html"],
                               ["download-ballot-box", "ballot-box.html"],
                               ["log-view", "log.html"]];
    $.each(permission_page_map, function(index, value) {
      var hide = true;
      for (var i = 0; i < value.length - 1; i++) {
        if ($.inArray(value[i], userContext['permissions']) !== -1) {
          hide = false;
          break;
        }
      }
      if (hide) {
        $('#side-menu').find('a[href~="' + value[1] + '"]').parent().hide();
      }
    });
  })
  .fail(function() {
    showErrorMessage('Viga kontekstiandmete laadimisel', false);
  });
}

/**
 * Show error message
 *
 * @param msg
 * @param {boolean} retain_content - Retain page content
 */
function showErrorMessage(msg, retain_content) {
  console.log(msg);
  if (!retain_content) {
    $('#page-wrapper').find('.row').hide();
  }
  $('#common-error-msg').find('p').html(msg);
  $('#common-error-msg').show();
}

function hideErrorMessage() {
  $('#common-error-msg').hide();
}

/**
 * Check user permissions
 *
 * @param {string} permission - Permission name to check
 * @return {boolean} Do user have permission
 */
function userHasPermission(permission) {
    return -1 !== $.inArray(permission, userContext['permissions']);
}

/**
 * Copy values from dictionary object to HTML element text nodes.
 *
 * @param {object} object_val - Dictionary object
 * @param {string} targetPrefix - Prefix for DOM id values
 */
function copyObjectToHtml(object_val, targetPrefix) {
  if ('undefined' === typeof(targetPrefix))
    targetPrefix = '';
  $.each(object_val, function(key, val) {
    $('#' + targetPrefix + key).html(val);
  });
}

function formatTime(dateTime, offset) {
  var seconds = dateTime.getSeconds();
  var minutes = dateTime.getMinutes();
  var hours = dateTime.getHours() + offset;

  seconds = seconds < 10 ? '0' + seconds : seconds;
  minutes = minutes < 10 ? '0' + minutes : minutes;
  hours = hours < 10 ? '0' + hours : hours;

  return hours + ":" + minutes + ":" + seconds;
}

// vim:sts=2 sw=2 et:
