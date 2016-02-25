var myApp = angular.module('myApp',[]);

$(document).ready(function(){
	// Set article height
 	var winH = $(window).height();
 	var footerH = $('#footer').height();
 	var headerH = $('#header').height();
 	$('.content article').css('min-height',winH-footerH-40+'px');
});

$(function(){
	$('#container').beforeAfter({
	animateIntro : true,
        introDelay : 1500,
        introDuration : 1000,
        showFullLinks : false
	});
});