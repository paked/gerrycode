class MainCtrl {
	constructor() {
		this.message = "Hello Feddie";

		if (window.token == "" || window.token == undefined) {
			$(location).attr("href", "#/login");
			return
		}
	}

}

export default MainCtrl;