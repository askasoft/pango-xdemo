(function($) {
	$(window).on('load', function() {
		$.ajax({
			url: $('#docx_dnload').attr('href'),
			xhr: function() {
				var xhr = new XMLHttpRequest();
				xhr.onreadystatechange = function() {
					if (xhr.readyState == 2) {
						if (xhr.status == 200) {
							xhr.responseType = "blob";
						} else {
							xhr.responseType = "text";
						}
					}
				};
				return xhr;
			},
			beforeSend: main.loadmask,
			success: function(data) {
				docx.renderAsync(data, $('#docx_preview')[0], $('#docx_styles')[0]);
			},
			error: main.ajax_error,
			complete: main.unloadmask
		});
	});
})(jQuery);
