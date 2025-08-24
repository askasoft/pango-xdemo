<s>[{{T .Loc "sitename"}}] Multi-factor authentication enrollment instructions</s>

<div>
	<p>Thank you for using {{T .Loc "sitename"}}.</p>
	<p>Use the authentication app on your mobile device to scan the QR code below.</p>
	<p><img src="data:image/png;base64,{{.QRCode}}"></p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
