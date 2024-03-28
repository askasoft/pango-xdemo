$(function() {
	function config_save() {
		var $f = $(this);
		$.ajax({
			url: './save',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: xmain.form_ajax_start($f),
			success: xmain.ajax_success,
			error: xmain.ajax_error,
			complete: xmain.form_ajax_end($f)
		});

		return false;
	}

	$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
		$($(e.target).attr('href')).find('textarea').autosize();
	});

	$('.cfgform').submit(config_save);
});
