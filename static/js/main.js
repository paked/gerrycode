app = angular.module('revco', ["ngRoute"]);

token = localStorage["token"]

app.config(['$routeProvider',
	function($routeProvider) {
		$routeProvider.
			when('/', {
				templateUrl: '/partials/home.html',
				controller: 'MainCtrl'
			}).
			when('/view/:project', {
				templateUrl: '/partials/view_project.html'
			}).
			when('/explore', {
				templateUrl: '/partials/explore.html'
			}).
			when('/make', {
				templateUrl: '/partials/make.html'
			}).
			when('/auth/:method', {
				templateUrl: '/partials/login.html',
				controller: 'AuthCtrl'
			}).
			when('/login', {
				redirectTo: '/auth/login'
			}).
			when('/register', {
				redirectTo: '/auth/register'
			}).
			otherwise({
				redirectTo: '/'
			});

	}]);

app.controller('MainCtrl', function($scope) {
	$scope.message = "PHONETAB EAT PHONETAB";
});

app.controller('AuthCtrl', function($scope, $routeParams, $http) {
	$scope.method = $routeParams["method"]

	$scope.other = function() {
		return $scope.method == "login" ? "register" : "login"
	}

	$scope.go = function() {
		var url = '/api/user/' + $scope.method + "?username=" + $scope.username + "&password=" + $scope.password + "&email=" + $scope.email;
		console.log(url);
		$http.post(url).
			success(function(data) {
				if (data.status.error) {
					$scope.error = data.status.message;
				}

				localStorage["token"] = data.data.token;
				token = localStorage["token"];
			}).
			error(function(data) {
				console.log(data);
			})
	}
})

app.controller('HeaderCtrl', function($scope, $location) {
	if (token == undefined) {
		$location.path("/login");
	}
});
