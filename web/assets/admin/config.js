$(function() {
	function config_save() {
		xdemo.loadmask();

		var $f = $(this);
		$.ajax({
			url: './save',
			type: 'POST',
			data: $f.serialize(),
			dataType: 'json',
			success: function(data, ts, xhr) {
				$.toast({
					icon: 'success',
					text: data.success
				});
			},
			error: xdemo.ajax_error,
			complete: function() {
				xdemo.unloadmask();
			}
		});

		return false;
	}

	$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
		$($(e.target).attr('href')).find('textarea').autosize();
	});

	$('.cfgform').submit(config_save);
});
