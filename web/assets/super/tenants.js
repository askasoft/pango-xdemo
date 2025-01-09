(function($) {
	var sskey = "a.tenants";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function tenants_reset() {
		$('#tenants_listform [name="p"]').val(1);
		$('#tenants_listform').formClear(true).submit();
		return false;
	}

	function tenants_search() {
		var $f = $('#tenants_listform');

		main.sssave(sskey, main.form_input_values($f));

		$.ajax({
			url: './list',
			method: 'POST',
			data: $f.serialize(),
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: main.list_builder($('#tenants_list')),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}


	function tenant_set_tr_values($tr, tenant) {
		main.set_table_tr_values($tr, tenant);
		$tr.find('td.domain > a').attr('href', '//' + tenant.name + '.' + main.domain).find('s').text(tenant.name + '.' + main.domain);
		main.blink($tr);
	}

	//----------------------------------------------------
	// create
	//
	function tenant_new() {
		$('#tenants_create_popup').popup('show');
		return false;
	}

	function tenant_create() {
		var $p = $('#tenants_create_popup').popup('update', { keyboard: false });

		$.ajax({
			url: './create',
			method: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var tenant = data.tenant;
				var $tb = $('#tenants_table > tbody'), $tr = $('#tenants_template tr').clone();

				$tr.attr({'id': 'tenant_' + tenant.name});
				$tr.find('td.check').append($('<input type="checkbox"/>').val(tenant.name));
				$tb.prepend($tr);

				tenant_set_tr_values($tr, tenant);
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// update
	//
	function tenant_edit() {
		var $tr = $(this).closest('tr'), $p = $('#tenants_update_popup');
		$p.find('[name=oname], [name=name]').val($tr.find('.name').text());
		$p.find('[name=comment]').val($tr.find('.comment > pre').text());
		$p.find('[name=name]').prop('readonly', $tr.data('current') || $tr.data('default') || false);
		$p.popup('show');
		return false;
	}

	function tenant_update() {
		var $p = $('#tenants_update_popup').popup('update', { keyboard: false });

		$.ajax({
			url: './update',
			method: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var tenant = data.tenant, $tr = $('#tenant_' + tenant.oname);

				tenant_set_tr_values($tr, tenant);
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// delete
	//
	function tenant_delete_confirm() {
		var name = $(this).closest('tr').find('.name').text();
		$('#tenant_delete_name').attr('placeholder', name).val('');
		$('#tenants_delete_popup').popup('show');
		return false;
	}

	function tenant_delete_check() {
		var $i = $(this);
		$i.closest('form').find('button[type=submit]').prop('disabled', $i.val() != $i.attr('placeholder'));
		return false;
	}

	function tenant_delete() {
		var $p = $('#tenants_delete_popup').popup('update', { keyboard: false });

		$.ajax({
			url: './delete',
			method: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				tenants_search();
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// init
	//
	function tenants_init() {
		main.list_init('tenants', sskey);
	
		$('#tenants_listform')
			.on('reset', tenants_reset)
			.on('submit', tenants_search)
			.submit();

		$('#tenants_list')
			.on('click', 'button.new', tenant_new)
			.on('click', 'button.edit', tenant_edit)
			.on('click', 'button.delete', tenant_delete_confirm);

		$('#tenants_create_popup')
			.on('submit', 'form', tenant_create)
			.on('click', '.ui-popup-footer button[type=submit]', tenant_create);
	
		$('#tenants_update_popup')
			.on('submit', 'form', tenant_update)
			.on('click', '.ui-popup-footer button[type=submit]', tenant_update);

		$('#tenants_delete_popup')
			.on('submit', 'form', tenant_delete)
			.on('click', '.ui-popup-footer button[type=submit]', tenant_delete);

		$('#tenant_delete_name').on('input', tenant_delete_check);
	}

	$(window).on('load', tenants_init);
})(jQuery);
