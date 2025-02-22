// yarn add -D gulp-newer gulp-svgmin

import gulp from 'gulp'
import newer from 'gulp-newer'
import svgmin from 'gulp-svgmin'

import env from './env.js'
import { browserSyncInstance } from './browserSync.js'

const path = {
  svg: 'assets/images/**/*.svg',
  other: 'assets/images/**/*.{png,jpg,ico,webp}',
  watch: 'assets/images/**/*.{png,jpg,svg,ico,webp}',
  favicon: 'assets/images/favicon/favicon.ico'
}

function img () {
  return gulp.src(path.other, { encoding: false })
    .pipe(newer(`${env.outputFolder}/statics/images`))
    .pipe(gulp.dest(`${env.outputFolder}/statics/images`))
}

function svg () {
  return gulp.src(path.svg, { encoding: false })
    .pipe(newer(`${env.outputFolder}/statics/images`))
    .pipe(svgmin())
    .pipe(gulp.dest(`${env.outputFolder}/statics/images`))
}

function favicon () {
  return gulp.src(path.favicon, { encoding: false })
    .pipe(gulp.dest(`${env.outputFolder}/`))
    .on('end', browserSyncInstance.reload)
}

export default {
  build: gulp.series(img, svg, favicon),
  watchPath: path.watch
}
