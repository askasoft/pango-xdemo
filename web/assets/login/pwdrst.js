(function($) {
	function send() {
		var $f = $('#pwdrst_form');
		
		$.ajax({
			url: './send',
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

	function init() {
		$('#pwdrst_form').on('submit', send);
	}

	$(window).on('load', init);
})(jQuery);
