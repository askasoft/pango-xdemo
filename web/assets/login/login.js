$(function() {
	function login() {
		var $f = $('#login_form');
		
		$.ajax({
			url: './login',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'success',
					text: data.success
				});

				setTimeout(function() {
					var origin = $f.find('input[name=origin]').val() || main.base || '/';
					location.href = origin;
				}, 500);
			},
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});
		return false;
	}

	$('#login_form').on('submit', login);
});
