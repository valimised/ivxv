function downloadBallot() {
	$('#loading').show();

	$.ajax({
		method: "POST",
		url: "/ivxv/cgi/download-ballot-box",
		success: function(url) {
			window.location.replace("/ivxv/data/ballot-box/" + url);
			$('#loading').hide();
		},
		error: function(jqXHR, textStatus, errorThrown) {
			$('#loading').hide();
			alert(errorThrown);
		}
	});
}
