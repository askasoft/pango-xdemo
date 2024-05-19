(function($) {
	function shell_exec() {
		var $f = $(this);

		if ($f.find('[name=command]').val().strip() == '') {
			return false;
		}

		var $sr = $('#shell_result').empty().hide();

		$.ajax({
			url: './exec',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				var lbls = $('#shell_map').data('labels');
				var $t = $('#shell_row').find('.row');
				for (var p in data) {
					var $r = $t.clone();
					$r.find('.ui-label').text(lbls[p] || p);
					$r.find('.ui-value').text(data[p]);
					$sr.append($r);
				}
				$sr.show();
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	function shell_init() {
		$('#shell_form').on('submit', shell_exec);
	}

	$(window).on('load', shell_init);
})(jQuery);

