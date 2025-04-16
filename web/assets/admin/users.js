(function($) {
	var sskey = "a.users";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function users_reset() {
		$('#users_listform [name="p"]').val(1);
		$('#users_listform').formClear(true).submit();
		return false;
	}

	function users_search(evt, callback) {
		var $f = $('#users_listform'), vs = main.form_input_values($f);

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
			success: main.list_builder($('#users_list'), callback),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}

	function users_prev_page(callback) {
		var pno = $('#users_list > .ui-pager > .pagination > .page-item.prev > a').attr('pageno');
		$('#users_listform input[name="p"]').val(pno);
		users_search(null, callback);
	}

	function users_next_page(callback) {
		var pno = $('#users_list > .ui-pager > .pagination > .page-item.next > a').attr('pageno');
		$('#users_listform input[name="p"]').val(pno);
		users_search(null, callback);
	}

	//----------------------------------------------------
	// export
	//
	function users_export() {
		$.ajaf({
			url: './export/csv',
			method: 'POST',
			data: $('#users_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// new
	//
	function user_new() {
		$('#users_detail_popup').popup({
			loaded: false,
			keyboard: false,
			ajax: {
				url: './new',
				method: 'GET'
			}
		}).popup('show', this);

		return false;
	}


	//----------------------------------------------------
	// detail
	//
	function user_id_from_tr($tr) {
		return $tr.attr('id').replace('user_', '');
	}

	function user_detail(action) {
		return user_detail_show(user_id_from_tr($(this).closest('tr')), action);
	}

	function user_detail_show(id, action) {
		$('#users_detail_popup').popup({
			loaded: false,
			keyboard: action == 'view',
			ajax: {
				url: action,
				method: 'GET',
				data: { id: id }
			}
		}).popup('show');

		return false;
	}

	function user_detail_action_trigger(selector, action) {
		$('#users_table > tbody > tr' + selector).find('button.' + action).trigger('click');
	}
	
	function user_detail_prev() {
		var action = $(this).attr('action');
		var id = $('#user_detail_id').val(), $tr = $('#user_' + id);

		$('#users_detail_popup').popup('hide');

		var $pv = $tr.prev('tr');
		if ($pv.length) {
			user_detail_show(user_id_from_tr($pv), action);
		} else {
			users_prev_page(user_detail_action_trigger.callback(':last-child', action));
		}
	}

	function user_detail_next() {
		var action = $(this).attr('action');
		var id = $('#user_detail_id').val(), $tr = $('#user_' + id);

		$('#users_detail_popup').popup('hide');

		var $nx = $tr.next('tr');
		if ($nx.length) {
			user_detail_show(user_id_from_tr($nx), action);
		} else {
			users_next_page(user_detail_action_trigger.callback(':first-child', action));
		}
	}

	function user_detail_submit() {
		$('#user_detail_id').val() == '0' ? user_create() : user_update();
		return false;
	}

	function user_detail_popup_loaded() {
		main.detail_popup_prevnext($('#users_detail_popup'), $('#users_list'), '#user_');
	}

	function user_detail_popup_shown() {
		main.detail_popup_shown($('#users_detail_popup'));
	}


	//----------------------------------------------------
	// update
	//
	var USM = $('#user_maps').data('status');
	var URM = $('#user_maps').data('role');

	function user_set_tr_values($tr, user) {
		main.set_table_tr_values($tr, user);
		$tr.attr('class', '').addClass(user.status);
		$tr.find('td.status').text(USM[user.status]);
		$tr.find('td.role').text(URM[user.role]);
		main.blink($tr);
	}

	function user_update() {
		var $p = $('#users_detail_popup');

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

				var user = data.user, $tr = $('#user_' + user.id);

				user_set_tr_values($tr, user);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}


	//----------------------------------------------------
	// create
	//
	function user_create() {
		var $p = $('#users_detail_popup');

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

				var user = data.user;
				var $tb = $('#users_table > tbody'), $tr = $('#users_template tr').clone();

				$tr.attr({'id': 'user_' + user.id});
				$tr.find('td.check').append($('<input type="checkbox"/>').val(user.id));
				$tb.prepend($tr);

				user_set_tr_values($tr, user);
				$tr.find('td.id, td.created_at').addClass('ro');
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}


	//----------------------------------------------------
	// deletes (selected / all)
	//
	function users_deletes(all) {
		var $p = $(all ? '#users_deleteall_popup' : '#users_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#users_table')).join(',');

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

				(all ? users_reset : users_search)();
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
	function users_deletebat() {
		var $p = $('#users_deletebat_popup').popup('update', { keyboard: false });
		var vs = main.form_input_values($p.find('form'));

		$.ajax({
			url: './deleteb',
			method: 'POST',
			data: $.param(vs, true),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				users_search();
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// updates (selected / all)
	//
	function users_updates() {
		var $p = $('#users_bulkedit_popup').popup({ keyboard: false });
		var ids = $p.find('[name=id]').val();

		$.ajax({
			url: './updates',
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

				var $trs = (ids == '*' ? $('#users_table > tbody > tr') : main.get_table_trs('#user_', ids.split(',')));

				var us = data.updates;
				if (us.status) {
					$trs.attr('class', '').addClass(us.status);
					$trs.find('td.status').text(USM[us.status]);
				}
				if (us.role) {
					$trs.find('td.role').text(URM[us.role]);
				}
				if ('cidr' in us) {
					$trs.find('td.cidr > pre').text(us.cidr);
				}
				main.blink($trs);
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
	function users_init() {
		main.list_init('users', sskey);

		$('#users_listform')
			.on('reset', users_reset)
			.on('submit', users_search)
			.submit();

		$('#users_export').on('click', users_export);
		$('#users_editsel').on('click', main.bulkedit_editsel_popup.callback('users'));
		$('#users_editall').on('click', main.bulkedit_editall_popup.callback('users'));

		$('#users_list')
			.on('click', 'button.new', user_new)
			.on('click', 'button.view', user_detail.callback("view"))
			.on('click', 'button.edit', user_detail.callback("edit"));

		$('#users_detail_popup')
			.on('loaded.popup', user_detail_popup_loaded)
			.on('shown.popup', user_detail_popup_shown)
			.on('keydown', main.detail_popup_keydown)
			.on('click', '.prev', user_detail_prev)
			.on('click', '.next', user_detail_next)
			.on('submit', 'form', user_detail_submit)
			.on('click', '.ui-popup-footer button[type=submit]', user_detail_submit);

		$('#users_deletesel_popup form').on('submit', users_deletes.callback(false));
		$('#users_deleteall_popup form').on('submit', users_deletes.callback(true));

		$('#users_deletebat_popup')
			.on('submit', 'form', users_deletebat)
			.on('click', '.ui-popup-footer button[type=submit]', users_deletebat);

		$('#users_bulkedit_popup')
			.on('change', '.col-form-label > input', main.bulkedit_label_click)
			.on('submit', 'form', users_updates)
			.on('click', '.ui-popup-footer button[type=submit]', users_updates);
	}

	$(window).on('load', users_init);
})(jQuery);
