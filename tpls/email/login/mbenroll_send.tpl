[{{HTML (T .Loc "site")}}] Multi-factor authentication enrollment instructions

<div>
	<p>Thank you for using {{T .Loc "site"}}.</p>
	<p>Use the authentication app on your mobile device to scan the QR code below.</p>
	<p><img src="data:image/png;base64,{{.QRCode}}"></p>
	<br>
	<p>Sincerely,</p>
	<p>The {{T .Loc "site"}} Team.</p>
</div>
