(function($) {
	var page = 1;
	if (location.hash) {
		var ps = $.queryParams(location.hash.substring(1));
		if (ps['page'] && parseInt(ps['page']) > 0) {
			page = parseInt(ps['page']);
		}
	}

	var on_pptx_rendered = function() {
		if (page > 1) {
			$('#pptx_preview > .slide').eq(page - 1).scrollIntoView();
		}
	};

	$(window).on('load', function() {
		$.ajax({
			url: $('#pptx_dnload').attr('href'),
			processData: false,
			xhr: function() {
				var xhr = new XMLHttpRequest();
				xhr.onreadystatechange = function() {
					if (xhr.readyState == 2) {
						if (xhr.status == 200) {
							xhr.responseType = "arraybuffer";
						} else {
							xhr.responseType = "text";
						}
					}
				};
				return xhr;
			},
			beforeSend: main.loadmask,
			success: function(data) {
				$('#pptx_preview').on('rendered.pptx', on_pptx_rendered).pptxToHtml({
					data: data
				});
			},
			error: main.ajax_error,
			complete: main.unloadmask
		});
	});
})(jQuery);
