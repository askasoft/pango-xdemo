(function($) {
	function configs_import() {
		var $p = $('#configs_import_popup').popup('update', { keyboard: false });

		$.ajaf({
			url: './import',
			method: 'POST',
			data: $p.find('form').serializeArray(),
			file: $p.find('input[type="file"]'),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success,
				});

				setTimeout(function() {
					location.reload();
				}, 3000);
			},
			error: main.ajax_error,
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});

		return false;
	}

	function configs_export() {
		$.ajaf({
			url: './export',
			method: 'POST',
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	function configs_save() {
		var $f = $(this);

		$.ajax({
			url: './save',
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

	function configs_init() {
		$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
			$($(e.target).attr('href')).find('textarea').autosize();
		});
	
		$('.cfgform').on('submit', configs_save);

		$('#configs_export').on('click', configs_export);
		$('#configs_import_popup').on('click', 'button[type=submit]', configs_import);
	}

	$(window).on('load', configs_init);
})(jQuery);
