(function($) {
	function send() {
		var $f = $(this);
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

	function exec() {
		var $f = $(this);
		$.ajax({
			url: location.href,
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
		$('#pwdrst_mail').on('submit', send);
		$('#pwdrst_exec').on('submit', exec);
	}

	$(window).on('load', init);
})(jQuery);
