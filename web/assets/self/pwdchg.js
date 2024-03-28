$(function() {
	function change() {
		var $f = $('#pwdchg_form');

		$.ajax({
			url: './change',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: xmain.form_ajax_start($f),
			success: xmain.ajax_success,
			error: xmain.form_ajax_error($f),
			complete: xmain.form_ajax_end($f)
		});
		return false;
	}

	$('#pwdchg_form').submit(change);
});
