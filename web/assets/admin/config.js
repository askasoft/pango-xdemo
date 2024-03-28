$(function() {
	function config_save() {
		var $f = $(this);
		$.ajax({
			url: './save',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: xdemo.form_ajax_start($f),
			success: xdemo.ajax_success,
			error: xdemo.ajax_error,
			complete: xdemo.form_ajax_end($f)
		});

		return false;
	}

	$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
		$($(e.target).attr('href')).find('textarea').autosize();
	});

	$('.cfgform').submit(config_save);
});
