<s>[{{T .Loc "sitename"}}] 多重身份验证注册指南</s>

<div>
	<p>感谢您使用 {{T .Loc "sitename"}}。</p>
	<p>请使用移动设备上的身份验证应用程序扫描下面的二维码。</p>
	<p><img src="data:image/png;base64,{{.QRCode}}"></p>
	<br>
	<p>此致,</p>
	<p>The {{T .Loc "sitename"}} Team.</p>
</div>
