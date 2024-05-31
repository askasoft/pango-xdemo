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
		var $f = $('#users_listform');

		main.sssave(sskey, main.form_input_values($f));

		$.ajax({
			url: './list',
			method: 'POST',
			data: $f.serialize(),
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: function(data) {
				main.list_build($('#users_list'), data);
				if (callback) {
					setTimeout(callback, 100);
				}
			},
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
	function user_detail(self, edit) {
		return user_detail_show($(self).closest('tr'), edit);
	}

	function user_detail_show($tr, edit) {
		var params = { id: $tr.attr('id').replace('user_', '') };

		$('#users_detail_popup').popup({
			loaded: false,
			keyboard: !edit,
			ajax: {
				url: edit ? "edit" : "view",
				method: 'GET',
				data: params
			}
		}).popup('show');

		return false;
	}

	function user_detail_prev() {
		var id = $('#user_detail_id').val(), $tr = $('#user_' + id);

		$('#users_detail_popup').popup('hide', $tr);

		var $pv = $tr.prev('tr');
		if ($pv.length) {
			user_detail_show($pv, $(this).attr('action') == 'edit');
		} else {
			users_prev_page(function() {
				$('#users_table > tbody > tr:last-child').find('button.edit').trigger('click');
			});
		}
	}

	function user_detail_next() {
		var id = $('#user_detail_id').val(), $tr = $('#user_' + id);

		$('#users_detail_popup').popup('hide', $tr);

		var $nx = $tr.next('tr');
		if ($nx.length) {
			user_detail_show($nx, $(this).attr('action') == 'edit');
		} else {
			users_next_page(function() {
				$('#users_table > tbody > tr:first-child').find('button.edit').trigger('click');
			});
		}
	}

	function user_detail_submit() {
		$('#user_detail_id').val() == '0' ? user_create() : user_update();
		return false;
	}

	function user_detail_popup_loaded() {
		$('#users_detail_popup')
			.find('form').on('submit', user_detail_submit).end()
			.find('.ui-popup-footer button[type=submit]').on('click', user_detail_submit);

		var $p = $('#users_detail_popup > .prev').on('click', user_detail_prev);
		var $n = $('#users_detail_popup > .next').on('click', user_detail_next);
		
		var id = $('#user_detail_id').val(), $tr = $('#user_' + id);
		var prev = $tr.prev('tr').length || $('#users_list > .ui-pager > .pagination > .page-item.prev.disabled').length == 0;
		var next = $tr.next('tr').length || $('#users_list > .ui-pager > .pagination > .page-item.next.disabled').length == 0;

		$p[(id != '0' && prev) ? 'show' : 'hide']();
		$n[(id != '0' && next) ? 'show' : 'hide']();
	}

	function user_detail_popup_shown() {
		$('#users_detail_popup')
			.find('.ui-popup-body').prop('scrollTop', 0).end()
			.find('[data-spy="niceSelect"]').niceSelect().end()
			.find('[data-spy="uploader"]').uploader().end()
			.find('input[type="text"]').textclear().end()
			.find('textarea').autosize().textclear().enterfire();
		$(window).trigger('resize');
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
				_token_: main.token,
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
	function users_deletebat(all) {
		var $p = $('#users_deletebat_popup').popup('update', { keyboard: false });

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

				if ($p.find('[name=id]').val() == '*') {
					users_search();
					return;
				}

				var us = data.updates;
				if (us) {
					var ids = main.get_table_checked_ids($('#users_table'));
					var $trs = main.get_table_trs('#user_', ids);

					if (us.status) {
						$trs.attr('class', '').addClass(us.status);
						$trs.find('td.status').text(USM[us.status]);
					}
					if (us.role) {
						$trs.find('td.role').text(URM[us.role]);
					}
					if (us.ucidr) {
						$trs.find('td.cidr > pre').text(us.cidr);
					}
					main.blink($trs);
				}
			},
			error: main.form_ajax_error($p),
			complete: function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}

	function users_editsel_click() {
		var ids = main.get_table_checked_ids($('#users_table'));
		$('#users_bulkedit_popup')
			.find('.editsel').show().end()
			.find('.editall').hide().end()
			.find('input[name=id]').val(ids.join(',')).end()
			.popup('show');
	}

	function users_editall_click() {
		$('#users_bulkedit_popup')
			.find('.editsel').hide().end()
			.find('.editall').show().end()
			.find('input[name=id]').val('*').end()
			.popup('show');
	}

	function users_bulkedit_input_change() {
		var $t = $(this);
		$t.parent().next().find(':input').prop('disabled', !$t.prop('checked'));
	}


	//----------------------------------------------------
	// init
	//
	function users_init() {
		if (!location.search) {
			$('#users_listform').formValues(main.ssload(sskey), true);
		}
		if (main.form_has_inputs($('#users_listform'))) {
			$('#users_listfset').fieldset('expand', 'show');
		}

		main.list_events('users');

		$('#users_listform')
			.on('reset', users_reset)
			.on('submit', users_search)
			.submit();

		$('#users_new').on('click', user_new);
		$('#users_export').on('click', users_export);
		$('#users_editsel').on('click', users_editsel_click);
		$('#users_editall').on('click', users_editall_click);

		$('#users_list')
			.on('click', 'button.view', function() { return user_detail(this, false); })
			.on('click', 'button.edit', function() { return user_detail(this, true); });

		$('#users_detail_popup')
			.on('loaded.popup', user_detail_popup_loaded)
			.on('shown.popup', user_detail_popup_shown);

		$('#users_deletesel_popup form').on('submit', function() { return users_deletes(false); });
		$('#users_deleteall_popup form').on('submit', function() { return users_deletes(true); });

		$('#users_deletebat_popup')
			.find('form').on('submit', users_deletebat).end()
			.find('.ui-popup-footer button[type=submit]').on('click', users_deletebat);

		$('#users_bulkedit_popup')
			.find('.col-form-label > input').on('change', users_bulkedit_input_change).end()
			.find('form').on('submit', users_updates).end()
			.find('.ui-popup-footer button[type=submit]').on('click', users_updates);
	}

	$(window).on('load', users_init);
})(jQuery);
