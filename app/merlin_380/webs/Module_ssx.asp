<!DOCTYPE html
	PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">

<head>
	<meta http-equiv="X-UA-Compatible" content="IE=Edge" />
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
	<meta HTTP-EQUIV="Pragma" CONTENT="no-cache" />
	<meta HTTP-EQUIV="Expires" CONTENT="-1" />
	<link rel="shortcut icon" href="images/favicon.png" />
	<link rel="icon" href="images/favicon.png" />
	<title>SSX</title>
	<link rel="stylesheet" type="text/css" href="index_style.css" />
	<link rel="stylesheet" type="text/css" href="form_style.css" />
	<link rel="stylesheet" type="text/css" href="usp_style.css" />
	<link rel="stylesheet" type="text/css" href="ParentalControl.css">
	<link rel="stylesheet" type="text/css" href="css/icon.css">
	<link rel="stylesheet" type="text/css" href="css/element.css">
	<script type="text/javascript" src="/state.js"></script>
	<script type="text/javascript" src="/popup.js"></script>
	<script type="text/javascript" src="/help.js"></script>
	<script type="text/javascript" src="/validator.js"></script>
	<script type="text/javascript" src="/js/jquery.js"></script>
	<script type="text/javascript" src="/general.js"></script>
	<script type="text/javascript" src="/switcherplugin/jquery.iphone-switch.js"></script>
	<script language="JavaScript" type="text/javascript" src="/client_function.js"></script>
	<script type="text/javascript" src="/dbconf?p=ssx_&v=<% uptime(); %>"></script>
	<script>
		var $j = jQuery.noConflict()
		function init() {
			show_menu(menu_hook)
			buildSwitch('enable-checkbox', 'ssx_enable')
			buildSwitch('ssl-checkbox', 'ssx_ssl')
		}

		function buildSwitch(id,key) {
			var $el = $j("#"+id)
			if (document.form[key].value != "1") {
				$el[0].checked = false
			} else {
				$el[0].checked = true
			}
			$j("#"+id).click(function () {
				document.form[key].value = $el[0].checked ? '1' : '0'
			})
		}

		function done_validating() {
				return true
		}

		function onSubmitCtrl(o, s) {
			showLoading(3)
			document.form.action_mode.value = s
			document.form.submit()
		}

		function reload_Soft_Center() {
			location.href = "/Main_Soft_center.asp"
		}

		function menu_hook(title, tab) {
			tabtitle[tabtitle.length - 1] = new Array("", "SSX")
			tablink[tablink.length - 1] = new Array("", "Module_ssx.asp")
		}
	</script>
</head>

<body onload="init()">
	<div id="TopBanner"></div>
	<div id="Loading" class="popup_bg"></div>
	<iframe name="hidden_frame" id="hidden_frame" src="" width="0" height="0" frameborder="0"></iframe>
	<form method="POST" name="form" action="/applydb.cgi?p=ssx_" target="hidden_frame">
		<input type="hidden" name="current_page" value="Module_ssx.asp" />
		<input type="hidden" name="next_page" value="Module_ssx.asp" />
		<input type="hidden" name="group_id" value="" />
		<input type="hidden" name="modified" value="0" />
		<input type="hidden" name="action_mode" value="" />
		<input type="hidden" name="action_script" value="" />
		<input type="hidden" name="action_wait" value="5" />
		<input type="hidden" name="first_time" value="" />
		<input type="hidden" name="preferred_lang" id="preferred_lang" value="<% nvram_get(" preferred_lang "); %>" />
		<input type="hidden" name="SystemCmd" onkeydown="onSubmitCtrl(this, ' Refresh ')" value="ssx.sh" />
		<input type="hidden" name="firmver" value="<% nvram_get(" firmver "); %>" />
		<input type="hidden" id="ssx_enable" name="ssx_enable" value='<% dbus_get_def("ssx_enable", "0"); %>' />
		<input type="hidden" id="ssx_ssl" name="ssx_ssl" value='<% dbus_get_def("ssx_ssl", "0"); %>' />
		<table class="content" align="center" cellpadding="0" cellspacing="0">
			<tr>
				<td width="17">&nbsp;</td>
				<td valign="top" width="202">
					<div id="mainMenu"></div>
					<div id="subMenu"></div>
				</td>
				<td valign="top">
					<div id="tabMenu" class="submenuBlock"></div>
					<table width="98%" border="0" align="left" cellpadding="0" cellspacing="0">
						<tr>
							<td align="left" valign="top">
								<table width="760px" border="0" cellpadding="5" cellspacing="0" bordercolor="#6b8fa3"
									class="FormTitle" id="FormTitle">
									<tr>
										<td bgcolor="#4D595D" colspan="3" valign="top">
											<div>&nbsp;</div>
											<div style="float:left;" class="formfonttitle">SSX</div>
											<div style="float:right; width:15px; height:25px;margin-top:10px">
												<img id="return_btn" onclick="reload_Soft_Center();" align="right"
													style="cursor:pointer;position:absolute;margin-left:-30px;margin-top:-25px;"
													title="Back to Software Center" src="/images/backprev.png"
													onMouseOver="this.src='/images/backprevclick.png'"
													onMouseOut="this.src='/images/backprev.png'"></img>
											</div>
											<table style="margin:10px 0px 0px 0px;" width="100%" border="1"
												align="center" cellpadding="4" cellspacing="0" bordercolor="#6b8fa3"
												class="FormTable" id="ssx_table">
												<thead>
													<tr>
														<td colspan="2">Options</td>
													</tr>
												</thead>
												<tr>
													<th>Enable</th>
													<td colspan="2">
														<div class="switch_field"
															style="display:table-cell;float: left;">
															<label for="enable-checkbox">
																<input id="enable-checkbox" class="switch" type="checkbox" style="display: none;">
																<div class="switch_container">
																	<div class="switch_bar"></div>
																	<div class="switch_circle transition_style">
																		<div></div>
																	</div>
																</div>
															</label>
														</div>
														<div id="ssx_version_show" style="padding-top:5px;margin-left:230px;margin-top:0px;">
															<i>Current version：<% dbus_get_def("ssx_version", "0.0.1"); %></i>
														</div>
													</td>
												</tr>
												<tr>
													<th>Server</th>
													<td colspan="2">
														<input type="text" maxlength="64" id="ssx_server" name="ssx_server" value='<% dbus_get_def("ssx_server", ""); %>' style="width:342px;float:left;" autocorrect="off" autocapitalize="off"/>
														&nbsp;
														<input id="ssl-checkbox" type="checkbox">
														<label for>SSL</label>
													</td>
												</tr>
												<tr>
													<th>DNS</th>
													<td colspan="2">
														<input type="text" maxlength="64" id="ssx_dns" name="ssx_dns" value='<% dbus_get_def("ssx_dns", ""); %>' style="width:342px;float:left;" autocorrect="off" autocapitalize="off"/>
													</td>
												</tr>
												<tr>
													<th>Doh Server</th>
													<td colspan="2">
														<input type="text" maxlength="64" id="ssx_doh_server" name="ssx_doh_server" value='<% dbus_get_def("ssx_doh_server", ""); %>' style="width:342px;float:left;" autocorrect="off" autocapitalize="off" placeholder="https://mozilla.cloudflare-dns.com/dns-query" />
													</td>
												</tr>
												<tr>
													<th>Blacklist Domains</th>
													<td colspan="2">
														<textarea id="ssx_blacklist_domains" name="ssx_blacklist_domains" style="width:342px;float:left;" rows="6" autocorrect="off" autocapitalize="off"><% dbus_get_def("ssx_blacklist_domains", ""); %></textarea>
													</td>
												</tr>
											</table>
											<div class="apply_gen">
												<button id="cmdBtn" class="button_gen" onclick="onSubmitCtrl(this, ' Refresh ')">Update</button>
											</div>
										</td>
									</tr>
								</table>
							</td>
							<td width="10" align="center" valign="top"></td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
	</form>
	</td>
	<div id="footer"></div>
</body>

</html>