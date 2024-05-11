$(function() {
	var sskey = "a.users";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function users_reset() {
		$('#users_listform [name="p"]').val(1);
		$('#users_listform').formClear(true).submit();
		return false;
	}

	function users_search() {
		var $f = $('#users_listform');

		main.sssave(sskey, main.form_input_values($f));

		$.ajax({
			url: './list',
			type: 'POST',
			data: $f.serialize(),
			beforeSend: function() {
				main.form_clear_invalid($f);
				main.loadmask();
			},
			success: main.list_builder($('#users_list')),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}

	if (!location.search) {
		$('#users_listform').formValues(main.ssload(sskey), true);
	}
	if (main.form_has_inputs($('#users_listform'))) {
		$('#users_listfset').fieldset('expand', 'show');
	}

	main.list_events('users');

	$('#users_listform').on('submit', users_search).submit();
	$('#users_listform').on('reset', users_reset);


	//----------------------------------------------------
	// export
	//
	function users_export() {
		$.ajaf({
			url: './export/csv',
			type: 'POST',
			data: $('#users_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}

	$('#users_export').on('click', users_export);


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

	$('#users_new').on('click', user_new);


	//----------------------------------------------------
	// detail
	//
	function user_detail(self, edit) {
		var $tr = $(self).closest('tr');
		var params = {
			id: $tr.attr('id').replace('user_', '')
		};

		$('#users_detail_popup').popup({
			loaded: false,
			keyboard: !edit,
			ajax: {
				url: edit ? "edit" : "view",
				method: 'GET',
				data: params
			}
		}).popup('show', this);

		return false;
	}

	function user_detail_submit() {
		$('#user_detail_id').val() == '0' ? user_create() : user_update();
		return false;
	}

	$('#users_list').on('click', 'button.view', function(evt) { return user_detail(this, false); });
	$('#users_list').on('click', 'button.edit', function(evt) { return user_detail(this, true); });

	$('#users_detail_popup')
		.on('loaded.popup', function() {
			$('#users_detail_popup')
				.find('form').on('submit', user_detail_submit).end()
				.find('.ui-popup-footer button[type=submit]').on('click', user_detail_submit);
		}).on('shown.popup', function() {
			$('#users_detail_popup')
				.find('.ui-popup-body').prop('scrollTop', 0).end()
				.find('[data-spy="niceSelect"]').niceSelect().end()
				.find('[data-spy="uploader"]').uploader().end()
				.find('input[type="text"]').textclear().end()
				.find('textarea').autosize().textclear().enterfire();
			$(window).trigger('resize');
		});


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
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
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
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
				$p.popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var user = data.user;
				var $tb = $('#users_table > tbody'), $tr = $tb.children('tr.template').clone();

				$tr.attr({ 'class': '', 'id': 'user_' + user.id});
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
			type: 'POST',
			data: {
				_token_: main.token,
				id: ids
			},
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
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

	$('#users_deletesel_popup form').on('submit', function() { return users_deletes(false); });
	$('#users_deleteall_popup form').on('submit', function() { return users_deletes(true); });


	//----------------------------------------------------
	// updates (selected / all)
	//
	function users_updates() {
		var $p = $('#users_bulkedit_popup').popup({ keyboard: false });

		$.ajax({
			url: './updates',
			type: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data, ts, xhr) {
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

	$('#users_editsel').on('click', function() {
		var ids = main.get_table_checked_ids($('#users_table'));
		$('#users_bulkedit_popup')
			.find('.editsel').show().end()
			.find('.editall').hide().end()
			.find('input[name=id]').val(ids.join(',')).end()
			.popup('show');
	});
	$('#users_editall').on('click', function() {
		$('#users_bulkedit_popup')
			.find('.editsel').hide().end()
			.find('.editall').show().end()
			.find('input[name=id]').val('*').end()
			.popup('show');
	});
	$('#users_bulkedit_popup')
		.find('.col-form-label > input').on('change', function() {
			var $t = $(this);
			$t.parent().next().find(':input').prop('disabled', !$t.prop('checked'));
		}).end()
		.find('form').on('submit', users_updates).end()
		.find('.ui-popup-footer button[type=submit]').on('click', users_updates);
});
