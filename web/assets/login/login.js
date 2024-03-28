$(function() {
	function login() {
		var $f = $('#login_form');
		
		$.ajax({
			url: './login',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: xmain.form_ajax_start($f),
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'success',
					text: data.success
				});

				setTimeout(function() {
					var origin = $f.find('input[name=origin]').val() || xmain.base || '/';
					location.href = origin;
				}, 500);
			},
			error: xmain.form_ajax_error($f),
			complete: xmain.form_ajax_end($f)
		});
		return false;
	}

	$('#login_form').submit(login);
});
