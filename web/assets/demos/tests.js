(function($) {
	function tests_invoke() {
		var $a = $(this), $s = $a.closest("#tests");

		$.ajax({
			url: $a.attr('href').substring(1),
			method: 'POST',
			beforeSend: $s.loadmask.delegate($s),
			success: function(data) {
				$.toast({
					icon: 'success',
					text: data
				});
			},
			error: main.ajax_error,
			complete: $s.unloadmask.delegate($s)
		});
		return false;
	}

	//----------------------------------------------------
	// init
	//
	function tests_init() {
		$('#tests a').on('click', tests_invoke);
	}

	$(window).on('load', tests_init);
})(jQuery);
