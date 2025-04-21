// yarn add -D sass gulp-sass gulp-autoprefixer gulp-group-css-media-queries gulp-clean-css gulp-header gulp-rename

import gulp from 'gulp'
import * as dartSass from 'sass'
import gulpSass from 'gulp-sass'
const sass = gulpSass(dartSass)
import autoPrefixer from 'gulp-autoprefixer'
import gcmq from 'gulp-group-css-media-queries'
import cleanCSS from 'gulp-clean-css'
import header from 'gulp-header'
import rename from 'gulp-rename'

import env from './env.js'
import { browserSyncInstance } from './browserSync.js'

const path = {
  pages: 'assets/styles/pages/*.sass',
  watch: 'assets/styles/**/*.sass',
}

const suffix = `-${env.hash}`

const styles = function () {
  if (env.production) {
    return gulp.src(path.pages)
      .pipe(header('@import "../variables"\n'))
      .pipe(sass().on('error', sass.logError))
      .pipe(autoPrefixer(['last 2 versions', '> 0.7%']))
      .pipe(gcmq())
      .pipe(cleanCSS({ level: 2 }))
      .pipe(rename({ suffix }))
      .pipe(gulp.dest(`${env.outputFolder}/statics/styles`))
  }
  return gulp.src(path.pages)
    .pipe(header('@import "../variables"\n'))
    .pipe(sass().on('error', sass.logError))
    .pipe(autoPrefixer(['last 2 versions', '> 0.7%']))
    .pipe(gcmq())
    .pipe(cleanCSS({
      level: 2,
      format: 'beautify'
    }))
    .pipe(gulp.dest(`${env.outputFolder}/statics/styles`))
    .on('end', () => browserSyncInstance.reload('*.css'))
}


export default {
  build: styles,
  watchPath: path.watch
}
