(function($) {
	function conv(url, $src, $des) {
		var $p = $('#chiconv');
		$.ajax({
			url: url,
			method: 'POST',
			data: { s: $src.val() },
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$des.val(data.success);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}

	function s2t() {
		return conv('./s2t', $('#simplified'), $('#traditional'));
	}

	function t2s() {
		return conv('./t2s', $('#traditional'), $('#simplified'));
	}

	//----------------------------------------------------
	// init
	//
	function chiconv_init() {
		$('#s2t').on('click', s2t);
		$('#t2s').on('click', t2s);
	}

	$(window).on('load', chiconv_init);
})(jQuery);
