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
			data: $f.serialize(),
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
	function pet_detail(self, action) {
		return pet_detail_show($(self).closest('tr'), action);
	}

	function pet_detail_show($tr, action) {
		var params = { id: $tr.attr('id').replace('pet_', '') };

		$('#pets_detail_popup').popup({
			loaded: false,
			keyboard: action == 'view',
			ajax: {
				url: action,
				method: 'GET',
				data: params
			}
		}).popup('show');

		return false;
	}

	function pet_detail_action_trigger(action) {
		$('#pets_table > tbody > tr:last-child').find('button.' + action).trigger('click');
	}

	function pet_detail_prev() {
		var action = $(this).attr('action');
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide', $tr);

		var $pv = $tr.prev('tr');
		if ($pv.length) {
			pet_detail_show($pv, action);
		} else {
			pets_prev_page(pet_detail_action_trigger.callback(action));
		}
	}

	function pet_detail_next() {
		var action = $(this).attr('action');
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide', $tr);

		var $nx = $tr.next('tr');
		if ($nx.length) {
			pet_detail_show($nx, action);
		} else {
			pets_next_page(pet_detail_action_trigger.callback(action));
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
		$.each(pet.habits, function(i, h) {
			hs.push($('<b>').text(PHM[h]));
		})
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
	function pets_deletebat(all) {
		var $p = $('#pets_deletebat_popup').popup('update', { keyboard: false });

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
					pets_search();
					return;
				}

				var us = data.updates;
				if (us) {
					var ids = main.get_table_checked_ids($('#pets_table'));
					var $trs = main.get_table_trs('#pet_', ids);

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
					main.blink($trs);
				}
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
		$('#pets_new').on('click', pet_new);
	
		$('#pets_list')
			.on('click', 'button.view', function() { return pet_detail(this, 'view'); })
			.on('click', 'button.edit', function() { return pet_detail(this, 'edit'); });

		$('#pets_detail_popup')
			.on('loaded.popup', pets_detail_popup_loaded)
			.on('shown.popup', pets_detail_popup_shown)
			.on('keydown', main.detail_popup_keydown)
			.on('click', '.prev', pet_detail_prev)
			.on('click', '.next', pet_detail_next)
			.on('submit', 'form', pet_detail_submit)
			.on('click', '.ui-popup-footer button[type=submit]', pet_detail_submit);

		$('#pets_deletesel_popup form').on('submit', function() { return pets_deletes(false); });
		$('#pets_deleteall_popup form').on('submit', function() { return pets_deletes(true); });

		$('#pets_deletebat_popup')
			.find('form').on('submit', pets_deletebat).end()
			.find('.ui-popup-footer button[type=submit]').on('click', pets_deletebat);

			$('#pets_editsel').on('click', main.bulkedit_editsel_click('pets'));
		$('#pets_editall').on('click', main.bulkedit_editall_click('pets'));

		$('#pets_bulkedit_popup')
			.find('.col-form-label > input').on('change', main.bulkedit_label_click).end()
			.find('form').on('submit', pets_updates).end()
			.find('.ui-popup-footer button[type=submit]').on('click', pets_updates);
	}

	$(window).on('load', pets_init);
})(jQuery);
