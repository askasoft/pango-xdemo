(function($) {
	var sskey = "d.pets";

	//----------------------------------------------------
	// list: pager & sorter
	//
	function pets_reset() {
		$('#pets_listform [name="p"]').val(1);
		$('#pets_listform').formClear(true).submit();
		return false;
	}

	function pets_search(evt, callback) {
		var $f = $('#pets_listform'), vs = main.form_input_values($f);

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
			success: main.list_builder($('#pets_list'), callback),
			error: main.form_ajax_error($f),
			complete: main.unloadmask
		});
		return false;
	}

	function pets_prev_page(callback) {
		var pno = $('#pets_list > .ui-pager > .pagination > .page-item.prev > a').attr('pageno');
		$('#pets_listform input[name="p"]').val(pno);
		pets_search(null, callback);
	}

	function pets_next_page(callback) {
		var pno = $('#pets_list > .ui-pager > .pagination > .page-item.next > a').attr('pageno');
		$('#pets_listform input[name="p"]').val(pno);
		pets_search(null, callback);
	}

	//----------------------------------------------------
	// export
	//
	function pets_export() {
		$.ajaf({
			url: './export/csv',
			method: 'POST',
			data: $('#pets_listform').serializeArray(),
			beforeSend: main.loadmask,
			error: main.ajax_error,
			complete: main.unloadmask
		});
		return false;
	}


	//----------------------------------------------------
	// new
	//
	function pet_new() {
		$('#pets_detail_popup').popup({
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
	function pet_id_from_tr($tr) {
		return $tr.attr('id').replace('pet_', '');
	}

	function pet_detail(action) {
		return pet_detail_show(pet_id_from_tr($(this).closest('tr')), action);
	}

	function pet_detail_show(id, action) {
		$('#pets_detail_popup').popup({
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

	function pet_detail_action_trigger(selector, action) {
		$('#pets_table > tbody > tr' + selector).find('button.' + action).trigger('click');
	}

	function pet_detail_prev() {
		var action = $(this).attr('action');
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide');

		var $pv = $tr.prev('tr');
		if ($pv.length) {
			pet_detail_show(pet_id_from_tr($pv), action);
		} else {
			pets_prev_page(pet_detail_action_trigger.callback(':last-child', action));
		}
	}

	function pet_detail_next() {
		var action = $(this).attr('action');
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide');

		var $nx = $tr.next('tr');
		if ($nx.length) {
			pet_detail_show(pet_id_from_tr($nx), action);
		} else {
			pets_next_page(pet_detail_action_trigger.callback(':first-child', action));
		}
	}

	function pet_detail_submit() {
		$('#pet_detail_id').val() == '0' ? pet_create() : pet_update();
		return false;
	}

	function pets_detail_popup_loaded() {
		main.detail_popup_prevnext($('#pets_detail_popup'), $('#pets_list'), '#pet_');
	}

	function pets_detail_popup_shown() {
		main.detail_popup_shown($('#pets_detail_popup'));
	}


	//----------------------------------------------------
	// update
	//
	var PGM = $('#pet_maps').data('gender');
	var POM = $('#pet_maps').data('origin');
	var PTM = $('#pet_maps').data('temper');
	var PHM = $('#pet_maps').data('habits');

	function pet_set_tr_values($tr, pet) {
		main.set_table_tr_values($tr, pet);
		$tr.find('td.gender').text(PGM[pet.gender]);
		$tr.find('td.origin').text(POM[pet.origin]);
		$tr.find('td.temper').text(PTM[pet.temper]);
		var hs = [];
		if (pet.habits) {
			for (var k in pet.habits) {
				if (pet.habits[k]) {
					hs.push($('<b>').text(PHM[k]));
				}
			}
		}
		$tr.find('td.habits').empty().append(hs);
		main.blink($tr);
	}

	function pet_update() {
		var $p = $('#pets_detail_popup');

		$.ajax({
			url: './update',
			method: 'POST',
			data: $p.find('form').serialize(),
			dataType: 'json',
			beforeSend: main.form_ajax_start($p),
			success: function(data) {
				$('#pets_detail_popup').popup('hide');

				$.toast({
					icon: 'success',
					text: data.success
				});

				var pet = data.pet, $tr = $('#pet_' + pet.id);

				pet_set_tr_values($tr, pet);
			},
			error: main.form_ajax_error($p),
			complete: main.form_ajax_end($p)
		});
		return false;
	}


	//----------------------------------------------------
	// create
	//
	function pet_create() {
		var $p = $('#pets_detail_popup');

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

				var pet = data.pet;
				var $tb = $('#pets_table > tbody'), $tr = $('#pets_template tr').clone();

				$tr.attr({'id': 'pet_' + pet.id});
				$tr.find('td.check').append($('<input type="checkbox"/>').val(pet.id));
				$tb.prepend($tr);

				pet_set_tr_values($tr, pet);
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
	function pets_deletes(all) {
		var $p = $(all ? '#pets_deleteall_popup' : '#pets_deletesel_popup').popup('update', { keyboard: false });
		var ids = all ? '*' : main.get_table_checked_ids($('#pets_table')).join(',');

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

				(all ? pets_reset : pets_search)();
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
	function pets_deletebat() {
		var $p = $('#pets_deletebat_popup').popup('update', { keyboard: false });
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

				pets_search();
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
	function pets_updates() {
		var $p = $('#pets_bulkedit_popup').popup('update', { keyboard: false });
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

				var $trs = (ids == '*' ? $('#pets_table > tbody > tr') : main.get_table_trs('#pet_', ids.split(',')));

				var us = data.updates;
				if (us.gender) {
					$trs.find('td.gender').text(PGM[us.gender]);
				}
				if (us.born_at) {
					$trs.find('td.born_at').text(main.format_date(us.born_at));
				}
				if (us.origin) {
					$trs.find('td.origin').text(POM[us.origin]);
				}
				if (us.temper) {
					$trs.find('td.temper').text(PTM[us.temper]);
				}
				if (us.habits) {
					var hs = [];
					$.each(us.habits, function(i, h) {
						hs.push($('<b>').text(PHM[h]));
					})
					$trs.find('td.habits').empty().append(hs);
				}
				if (us.updated_at) {
					$trs.find('td.updated_at').text(main.format_time(us.updated_at));
				}
				main.blink($trs);
			},
			error: main.form_ajax_error($p),
			complete:  function() {
				$p.unloadmask().popup('update', { keyboard: true });
			}
		});
		return false;
	}


	//----------------------------------------------------
	// init
	//
	function pets_init() {
		main.list_init('pets', sskey);
	
		$('#pets_listform')
			.on('reset', pets_reset)
			.on('submit', pets_search)
			.submit();

		$('#pets_export').on('click', pets_export);
	
		$('#pets_list')
			.on('click', 'button.new', pet_new)
			.on('click', 'button.view', pet_detail.callback('view'))
			.on('click', 'button.edit', pet_detail.callback('edit'));

		$('#pets_detail_popup')
			.on('loaded.popup', pets_detail_popup_loaded)
			.on('shown.popup', pets_detail_popup_shown)
			.on('keydown', main.detail_popup_keydown)
			.on('click', '.prev', pet_detail_prev)
			.on('click', '.next', pet_detail_next)
			.on('submit', 'form', pet_detail_submit)
			.on('click', '.ui-popup-footer button[type=submit]', pet_detail_submit);

		$('#pets_deletesel_popup form').on('submit', pets_deletes.callback(false));
		$('#pets_deleteall_popup form').on('submit', pets_deletes.callback(true));

		$('#pets_deletebat_popup')
			.on('submit', 'form', pets_deletebat)
			.on('click', '.ui-popup-footer button[type=submit]', pets_deletebat);

		$('#pets_editsel').on('click', main.bulkedit_editsel_popup.callback('pets'));
		$('#pets_editall').on('click', main.bulkedit_editall_popup.callback('pets'));

		$('#pets_bulkedit_popup')
			.on('change', '.col-form-label > input', main.bulkedit_label_click)
			.on('submit', 'form', pets_updates)
			.on('click', '.ui-popup-footer button[type=submit]', pets_updates);
	}

	$(window).on('load', pets_init);
})(jQuery);
