myApp.controller('myCtrl', function ($scope) {
    
    $scope.hash =  window.location.hash.substr(1);
    if(!$scope.hash){
        $scope.hash = 'mulifs--music-library-filesystem';
        $scope.boolChangeClass = false;
    }
    else{
        $scope.boolChangeClass = false;
    }
    $('.menu li a').click(function(){
        $scope.boolChangeClass=true;
    });
});

myApp.directive("scroll", function ($window) {
    return function(scope, element, attrs) {
        angular.element($window).bind("scroll", function() {
             if (this.pageYOffset >= 339) {
                 scope.boolChangeClass = true;
             } else {
                 scope.boolChangeClass = false;
             }
             if (this.pageYOffset >= 310) {
                 scope.fadeoutThis = true;
             } else {
                 scope.fadeoutThis = false;
             }
            scope.$apply();
        });
    };
});