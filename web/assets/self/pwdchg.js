$(function() {
	function change() {
		var $f = $('#pwdchg_form');

		$.ajax({
			url: './change',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: main.ajax_success,
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});
		return false;
	}

	$('#pwdchg_form').submit(change);
});
