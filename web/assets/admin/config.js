$(function() {
	function config_save() {
		var $f = $(this);

		$.ajax({
			url: './save',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($f),
			success: main.ajax_success,
			error: main.ajax_error,
			complete: main.form_ajax_end($f)
		});

		return false;
	}

	$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
		$($(e.target).attr('href')).find('textarea').autosize();
	});

	$('.cfgform').on('submit', config_save);
});
