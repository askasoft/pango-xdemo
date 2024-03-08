$(function() {
	function login() {
		$('#login_form').loadmask();

		$.ajax({
			url: './login',
			type: 'POST',
			data: $('#login_form').serialize(),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'success',
					text: data.success,
					hideAfter: 3000
				});

				setTimeout(function() {
					var origin = $('#login_form input[name=origin]').val() || xdemo.base || '/';
					location.href = origin;
				}, 500);
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#login_form').unloadmask();
			}
		});
		return false;
	}

	$('#login_form').submit(login);
});
