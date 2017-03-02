var gulp = require('gulp');
var htmlmin = require('gulp-htmlmin');
var fileinclude  = require('gulp-file-include');

gulp.task('fileinclude', function() {
    gulp.src(['src/**/*.html','!src/include/**.html'])
        .pipe(fileinclude({
          prefix: '@@',
          basepath: '@file'
        }))
        .pipe(htmlmin({
            removeComments: true,
            collapseWhitespace: true,
            collapseBooleanAttributes: true,
            removeEmptyAttributes: true,
            removeScriptTypeAttributes: true,
            removeStyleLinkTypeAttributes: true,
            minifyJS: true,
            minifyCSS: true
        }))
    .pipe(gulp.dest('.'));
});
gulp.task('default',['fileinclude']); 
