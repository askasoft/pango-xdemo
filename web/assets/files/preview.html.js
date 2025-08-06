(function($) {
	$(window).on('load resize', function() {
		var $i = $('#html_iframe');
		$i.css({ width: '100%', height: $i[0].contentWindow.document.documentElement.scrollHeight + 'px' });
	});
})(jQuery);
