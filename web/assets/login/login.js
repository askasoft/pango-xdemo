(function($) {
	function login() {
		var $f = $('#login_form'), $la = $('#login_alert').empty();
		
		$.ajax({
			url: './login',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: function(data) {
				if (data.mfa) {
					main.show_alert($la, 'primary', data.message);
					$f.find('input[name=passcode]').closest('.row').removeClass('hidden');
					$f.find('.desc.'+data.mfa).removeClass('hidden');
					return;
				}

				if (data.success) {
					$.toast({
						icon: 'success',
						text: data.success
					});
	
					setTimeout(function() {
						var origin = $f.find('input[name=origin]').val() || main.base || '/';
						location.href = origin;
					}, 500);
				}
			},
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});
		return false;
	}

	function login_mfa_enroll() {
		$('#login_mfa_enroll_popup').popup('hide');

		var $f = $('#login_form');
		$.ajax({
			url: './mfa_enroll',
			method: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: main.ajax_success,
			error: main.form_ajax_error($f),
			complete: main.form_ajax_end($f)
		});
		return false;
	}

	function init() {
		$('#login_form').on('submit', login);
		$('#login_mfa_enroll_popup form').on('submit', login_mfa_enroll);
	}

	$(window).on('load', init);
})(jQuery);
