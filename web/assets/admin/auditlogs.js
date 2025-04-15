(function($) {
	var sskey = "a.auditlogs";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function auditlogs_reset() {
		$('#auditlogs_listform [name="p"]').val(1);
		$('#auditlogs_listform').formClear(true).submit();
		return false;
	}

	function auditlogs_search(evt, callback) {
		var $f = $('#auditlogs_listform'), vs = main.form_input_values($f);

		main.sssave(sskey, vs);
		main.location_replace_search(vs);

		$.ajax({
			url: './list',
			method: 'POST',
			data: $.param(vs, true),
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: main.list_builder($('#auditlogs_list'), callback),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// export
	//
	function auditlogs_export() {
		$.ajaf({
			url: './export/csv',
			method: 'POST',
			data: $('#auditlogs_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (selected / all)
	//
	function auditlogs_deletes(all) {
		var $p = $(all ? '#auditlogs_deleteall_popup' : '#auditlogs_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#auditlogs_table')).join(',');

		$.ajax({
			url: './deletes',
			method: 'POST',
			data: {
				id: ids
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				(all ? auditlogs_reset : auditlogs_search)();
			},
			error: main.ajax_error,
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (batch)
	//
	function auditlogs_deletebat() {
		var $p = $('#auditlogs_deletebat_popup').popup('update', { keyboard: false });

		$.ajax({
			url: './deleteb',
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

				auditlogs_search();
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
	function auditlogs_init() {
		main.list_init('auditlogs', sskey);

		$('#auditlogs_listform')
			.on('reset', auditlogs_reset)
			.on('submit', auditlogs_search)
			.submit();

		$('#auditlogs_export').on('click', auditlogs_export);

		$('#auditlogs_deletesel_popup form').on('submit', auditlogs_deletes.callback(false));
		$('#auditlogs_deleteall_popup form').on('submit', auditlogs_deletes.callback(true));

		$('#auditlogs_deletebat_popup')
			.on('submit', 'form', auditlogs_deletebat)
			.on('click', '.ui-popup-footer button[type=submit]', auditlogs_deletebat);
	}

	$(window).on('load', auditlogs_init);
})(jQuery);
