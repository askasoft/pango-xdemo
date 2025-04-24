(function($) {
	function sql_exec() {
		var $f = $(this);

		if ($f.find('[name=sql]').val().strip() == '') {
			return false;
		}

		var $srs = $('#sql_results').empty().hide();

		$.ajax({
			url: './exec',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				$srs.show();

				var $t = $('#sql_result').children('li');

				$.each(data, function(i, d) {
					var $sr = $t.clone();

					$sr.find('.sql').text(d.sql);
					$sr.find('.err').text(d.error).toggle(!!d.error);
					$sr.find('.res').text((d.elapsed ? 'Elapsed: ' + d.elapsed : '') + (d.effected ? '\nEffected: ' + d.effected : ''));

					if (d.columns) {
						var $th = $sr.find('thead'), $tr = $('<tr>');
						$tr.append($('<th class="no">').text('##'));
						$.each(d.columns, function(i, c) {
							$tr.append($('<th class="col">').text(c));
						});
						$th.append($tr);
					}
					if (d.datas) {
						var $tb = $sr.find('tbody');
						$.each(d.datas, function(i, r) {
							var $tr = $('<tr>');
							$tr.append($('<td class="no">').text('#' + (i+1)));
							$.each(r, function(i, c) {
								$tr.append($('<td class="col">').text(c));
							});
							$tb.append($tr);
						});
					}

					$srs.append($sr);
				});
			},
			error: main.ajax_error,
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	function sql_init() {
		$('#sql_form').on('submit', sql_exec);
	}

	$(window).on('load', sql_init);
})(jQuery);

