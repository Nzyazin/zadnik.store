// yarn add -D gulp-pug gulp-data gulp-plumber gulp-typograf gulp-format-html gulp-htmlmin gulp-rename gulp-header

import gulp from 'gulp'
import pug from 'gulp-pug'
import data from "gulp-data"
import plumber from 'gulp-plumber'
import typograf from 'gulp-typograf'
import formatHtml from 'gulp-format-html'
import htmlmin from 'gulp-htmlmin'
import rename from 'gulp-rename'
import header from 'gulp-header'

import env from './env.js'
import { browserSyncInstance } from './browserSync.js'

const path = {
  pages: 'assets/views/pages/*.pug',
  watch: 'assets/views/**/*.pug',
  error: 'assets/views/pages/error.pug',
  siteMap: 'assets/views/pages/_site-map.pug'
}

const typografConfig = {
  locale: ['ru', 'en-US'],
  safeTags: [
    ['<head>', '</head>']
  ]
}

const htmlminConfig = { collapseWhitespace: true, conservativeCollapse: true, minifyJS: true, minifyCSS: true }

function dataView (file) {
  return {
    VIEW: file.stem,
    PRODUCTION: env.production,
    HASH: env.hash,
    URL: env.url,
    DOMAIN: env.domain,
  }
}

function view () {
  if (env.production) {
    return gulp.src(path.pages, { ignore: [path.error, path.siteMap] })
      .pipe(plumber())
      .pipe(data(dataView))
      .pipe(pug())
      .pipe(htmlmin(htmlminConfig))
      .pipe(typograf(typografConfig))
      .pipe(gulp.dest(env.outputFolder))
  }
  return gulp.src(path.pages)
    .pipe(plumber())
    .pipe(data(dataView))
    .pipe(pug())
    .pipe(formatHtml())
    .pipe(typograf(typografConfig))
    .pipe(gulp.dest(env.outputFolder))
    .on('end', browserSyncInstance.reload)
}

function sitemap () {
  return gulp.src(path.siteMap)
    .pipe(plumber())
    .pipe(data(dataView))
    .pipe(pug())
    .pipe(rename({
      basename: 'sitemap',
      extname: ".xml"
    }))
    .pipe(gulp.dest(env.outputFolder))
}

function error () {
  return gulp.src(path.error)
    .pipe(plumber())
    .pipe(data(dataView))
    .pipe(pug())
    .pipe(htmlmin(htmlminConfig))
    .pipe(typograf(typografConfig))
    .pipe(header('<?php header($_SERVER[\'SERVER_PROTOCOL\']." 404 Not Found");?>'))
    .pipe(rename({ extname: '.php' }))
    .pipe(gulp.dest(env.outputFolder))
}

export default {
  build: env.production ? gulp.series(view, sitemap, error) : view,
  watchPath: path.watch
}
