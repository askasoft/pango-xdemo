$(function() {
	function change() {
		$('#pwdchg_form').loadmask();

		$.ajax({
			url: './change',
			type: 'POST',
			data: $('#pwdchg_form').serialize(),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'success',
					text: data.success
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				$('#pwdchg_form').unloadmask();
			}
		});
		return false;
	}

	$('#pwdchg_form').submit(change);
});
