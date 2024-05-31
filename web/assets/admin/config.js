(function($) {
	function config_save() {
		var $f = $(this);

		$.ajax({
			url: './save',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: main.ajax_success,
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	function config_init() {
		$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
			$($(e.target).attr('href')).find('textarea').autosize();
		});
	
		$('.cfgform').on('submit', config_save);
	}

	$(window).on('load', config_init);
})(jQuery);
