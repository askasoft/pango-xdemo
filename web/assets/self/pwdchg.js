$(function() {
	function change() {
		var $f = $('#pwdchg_form');

		$.ajax({
			url: './change',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: xdemo.form_ajax_start($f),
			success: xdemo.ajax_success,
			error: xdemo.form_ajax_error($f),
			complete: xdemo.form_ajax_end($f)
		});
		return false;
	}

	$('#pwdchg_form').submit(change);
});
