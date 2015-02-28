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

app.controller('AuthCtrl', function($scope, $routeParams, $http, $location) {
	$scope.method = $routeParams["method"]
	$scope.email = ""

	$scope.other = function() {
		return $scope.method == "login" ? "register" : "login"
	}

	$scope.go = function() {
		var url = '/api/user/' + $scope.method + "?username=" + $scope.username + "&password=" + $scope.password + "&email=" + $scope.email;
		console.log(url);
		$http.post(url).
			success(function(data) {
				console.log(data)
				if (data.status.error) {
					$scope.error = data.message;
					return
				}
				
				if (data.data == undefined) {
					$location.path('/login')
					return
				}

				localStorage["token"] = data.data;
				token = localStorage["token"];
				$location.path("/")
			}).
			error(function(data) {
				console.log(data);
			})
	}
})

app.controller('HeaderCtrl', function($scope, $location, $http) {
	function checkAuth() {
		$location.path("/login");
		$scope.loggedIn = false;
	}

	$scope.loggedIn = true;
	if (token === undefined || token == "undefined") {
		checkAuth();
		return
	}

	$http.get('/api/user?access_token=' + token).
		success(function(data) {
			if (data.status.error) {
				checkAuth();
				return
			}

			$scope.user = data.data;
		});
});
