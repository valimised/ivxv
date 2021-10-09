/*
 * IVXV Internet voting framework
 *
 * Administrator interface - Ballot box download page
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

      if ((state.collector_state == 'NOT INSTALLED') ||
        (state.collector_state == 'INSTALLED')) {
        $('#common-warning-msg').show('slow');
        $('#common-warning-msg .panel-body p').text('Kogumisteenus pole seadistatud. Valimiskasti allalaadimine pole võimaik.');
        $('#panel-download-form').hide();
      } else if (state.election.phase == 'PREPARING') {
        $('#common-warning-msg').show('slow');
        $('#common-warning-msg .panel-body p').text('Hääletamine on ettevalmistamise faasis. Valimiskasti allalaadimine pole võimaik.');
        $('#panel-download-form').hide();
      } else if (state.election.phase != 'FINISHED') {
        $('#common-warning-msg').show('slow');
        $('#common-warning-msg .panel-body p').text('Hääletusperioodil on võimalik väljastada vaid e-valimiskasti varukoopiat.');
        $('#panel-download-form').show();
      } else {
        $('#common-warning-msg').hide('slow');
        $('#panel-download-form').show();
      }
    })
    .fail(function() {
      $('#loadstatus')
        .removeClass('text-info')
        .addClass('text-danger')
        .html('Viga andmete laadimisel: ' + formatTime(loadDate, 0));
      showErrorMessage('Viga seisundi laadimisel', true);
    });

  // load ballot box state
  $.getJSON('cgi/ballot-box-state', function(state) {
      $('#panel-download-ballot-box').remove();

      if (state.data.length === 0) {
        console.log('There is no ballot box created for download');
        return;
      }

      var panel = $('#panel-download-ballot-box-template').clone().prop({
        id: 'panel-download-ballot-box'
      });
      panel.find('.list-group').empty();

      $(state.data).each(function() {
        var line = null;
        var message = this.filename.search('consolidated') === -1 ?
          'Eksporditud e-valimiskast' :
          'Eksporditud ja konsolideeritud e-valimiskast';
        if (this.state == 'ready') {
          line = $('#panel-download-ballot-box-template .list-group-item-success')
            .clone()
            .attr('href', '/ivxv/data/ballot-box/' + this.filename);
        } else if (this.state == 'prepare') {
          line = $('#panel-download-ballot-box-template .list-group-item-info')
            .clone();
          line.find('pre').text(this.log).show();
        } else {
          line = $('#panel-download-ballot-box-template .list-group-item-danger')
            .clone();
          message = 'Unknown ballot box state: ' + this.state;
        }

        line.find('span').first().text(message);
        line.find('span em').text(this.timestamp);
        panel.find('.list-group').append(line);
      });

      $('#panel-download-ballot-box-template').after(panel);
      panel.show();
    })
    .fail(function(jqXHR, textStatus) {
      $('#panel-download-ballot-box').remove();
      var panel = $('#panel-download-ballot-box-template').clone().prop({
        id: 'panel-download-ballot-box'
      });
      panel.find('.list-group').empty();
      $('#panel-download-ballot-box-template').after(panel);

      var line = $('#panel-download-ballot-box-template .list-group-item-danger')
        .clone();
      line
        .find('span')
        .first()
        .text(
          'Serveri viga allalaadimiseks ettevalmistatud valimiskastide olekuandmete laadimisel');
      line.find('span em').remove();

      panel.find('.list-group').append(line);
      panel.show();
    });
}

/**
 * Download ballot box
 */
function downloadBallot(consolidate) {
  $('#loading').show();
  url = (
    consolidate === 1 ?
    '/ivxv/cgi/download-consolidated-ballot-box' :
    '/ivxv/cgi/download-ballot-box')

  $.ajax({
    method: 'POST',
    url: url,

    success: function(url) {
      $('#loading').hide();
      $('#panel-download-form .panel-body')
        .text('Server alustas valimiskasti ettevalmistamist allalaadimiseks');
    },

    error: function(jqXHR, textStatus, errorThrown) {
      $('#loading').hide();
      alert(errorThrown);
    }

  });
}
