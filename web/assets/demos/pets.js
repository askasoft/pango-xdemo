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
	function pet_detail(self, edit) {
		return pet_detail_show($(self).closest('tr'), edit);
	}

	function pet_detail_show($tr, edit) {
		var params = { id: $tr.attr('id').replace('pet_', '') };

		$('#pets_detail_popup').popup({
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

	function pet_detail_prev() {
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide', $tr);

		var $pv = $tr.prev('tr');
		if ($pv.length) {
			pet_detail_show($pv, $(this).attr('action') == 'edit');
		} else {
			pets_prev_page(function() {
				$('#pets_table > tbody > tr:last-child').find('button.edit').trigger('click');
			});
		}
	}

	function pet_detail_next() {
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);

		$('#pets_detail_popup').popup('hide', $tr);

		var $nx = $tr.next('tr');
		if ($nx.length) {
			pet_detail_show($nx, $(this).attr('action') == 'edit');
		} else {
			pets_next_page(function() {
				$('#pets_table > tbody > tr:first-child').find('button.edit').trigger('click');
			});
		}
	}

	function pet_detail_submit() {
		$('#pet_detail_id').val() == '0' ? pet_create() : pet_update();
		return false;
	}

	function pets_detail_popup_loaded() {
		$('#pets_detail_popup')
			.find('form').on('submit', pet_detail_submit).end()
			.find('.ui-popup-footer button[type=submit]').on('click', pet_detail_submit);

		var $p = $('#pets_detail_popup > .prev').on('click', pet_detail_prev);
		var $n = $('#pets_detail_popup > .next').on('click', pet_detail_next);
		
		var id = $('#pet_detail_id').val(), $tr = $('#pet_' + id);
		var prev = $tr.prev('tr').length || $('#pets_list > .ui-pager > .pagination > .page-item.prev.disabled').length == 0;
		var next = $tr.next('tr').length || $('#pets_list > .ui-pager > .pagination > .page-item.next.disabled').length == 0;

		$p[(id != '0' && prev) ? 'show' : 'hide']();
		$n[(id != '0' && next) ? 'show' : 'hide']();
	}

	function pets_detail_popup_shown() {
		$('#pets_detail_popup')
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

	function pets_editsel_click() {
		var ids = main.get_table_checked_ids($('#pets_table'));
		$('#pets_bulkedit_popup')
			.find('.editsel').show().end()
			.find('.editall').hide().end()
			.find('input[name=id]').val(ids.join(',')).end()
			.popup('show');
	}

	function pets_editall_click() {
		$('#pets_bulkedit_popup')
			.find('.editsel').hide().end()
			.find('.editall').show().end()
			.find('input[name=id]').val('*').end()
			.popup('show');
	}

	function pets_bulkedit_input_change() {
		var $t = $(this), c = $t.prop('checked');
		var $i = $t.parent().next().find(':input').prop('disabled', !c);
		if ($t.data('niceselect')) {
			$i.niceSelect('update');
		}
	}


	//----------------------------------------------------
	// init
	//
	function pets_init() {
		if (!location.search) {
			$('#pets_listform').formValues(main.ssload(sskey), true);
		}
		if (main.form_has_inputs($('#pets_listform'))) {
			$('#pets_listfset').fieldset('expand', 'show');
		}

		main.list_events('pets');
	
		$('#pets_listform')
			.on('reset', pets_reset)
			.on('submit', pets_search)
			.submit();

		$('#pets_export').on('click', pets_export);
		$('#pets_new').on('click', pet_new);
	
		$('#pets_list')
			.on('click', 'button.view', function() { return pet_detail(this, false); })
			.on('click', 'button.edit', function() { return pet_detail(this, true); });

		$('#pets_detail_popup')
			.on('loaded.popup', pets_detail_popup_loaded)
			.on('shown.popup', pets_detail_popup_shown);

		$('#pets_deletesel_popup form').on('submit', function() { return pets_deletes(false); });
		$('#pets_deleteall_popup form').on('submit', function() { return pets_deletes(true); });

		$('#pets_deletebat_popup')
			.find('form').on('submit', pets_deletebat).end()
			.find('.ui-popup-footer button[type=submit]').on('click', pets_deletebat);

			$('#pets_editsel').on('click', pets_editsel_click);
		$('#pets_editall').on('click', pets_editall_click);

		$('#pets_bulkedit_popup')
			.find('.col-form-label > input').on('change', pets_bulkedit_input_change).end()
			.find('form').on('submit', pets_updates).end()
			.find('.ui-popup-footer button[type=submit]').on('click', pets_updates);
	}

	$(window).on('load', pets_init);
})(jQuery);
