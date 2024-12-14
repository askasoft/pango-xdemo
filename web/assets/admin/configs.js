(function($) {
	function toggle_login_method() {
		var saml = $('input[name="secure_login_method"]:checked').val() == 'S';
		$('[name="secure_login_mfa"]').prop('disabled', saml).closest('.row')[saml ? 'slideUp' : 'slideDown']();
		$('[name^="secure_saml_"]').prop('disabled', !saml).closest('.row')[saml ? 'slideDown' : 'slideUp']();
	}

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

	function configs_tab_show(id) {
		var $t = $('a[href="#' + id + '"]');
		if ($t.length) {
			new bootstrap.Tab($t.get(0)).show();
		}
	}

	function configs_init() {
		$('a[data-bs-toggle="tab"]').on('shown.bs.tab', function(e) {
			var t = $(e.target).attr('href');
			$(t).find('textarea').autosize();
			history.replaceState(null, null, location.href.split('#')[0] + t);
		});

		$('.cfgform').find('textarea').autosize();
		$('.cfgform').on('submit', configs_save);

		$('#configs_export').on('click', configs_export);
		$('#configs_import_popup').on('click', 'button[type=submit]', configs_import);

		var cg = location.href.substrAfter('#'), cc = cg;
		if (cc.startsWith('cg_')) {
			cc = $('#' + cc).parent().closest('.tab-pane').attr('id');
		}
		if (cc.startsWith('cc_')) {
			configs_tab_show(cc);
		}
		if (cg.startsWith('cg_')) {
			configs_tab_show(cg);
		}

		$('.cfgform input[name="secure_login_method"]').change(toggle_login_method);
		toggle_login_method();
	}

	$(window).on('load', configs_init);
})(jQuery);
