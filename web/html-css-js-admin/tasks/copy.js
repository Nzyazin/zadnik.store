// yarn add -D gulp-plumber gulp-pug gulp-data gulp-rename

import gulp from 'gulp'
import plumber from 'gulp-plumber'
import pug from 'gulp-pug'
import data from "gulp-data"
import rename from 'gulp-rename'

import env from './env.js'

const dataView = { URL: env.url }

function htaccess () {
  return gulp.src('assets/files/.htaccess.pug', { encoding: false })
    .pipe(plumber())
    .pipe(data(dataView))
    .pipe(pug())
    .pipe(rename({ extname: '' }))
    .pipe(gulp.dest(`${env.outputFolder}`))
}

function robots () {
  return gulp.src('assets/files/robots.pug', { encoding: false })
    .pipe(plumber())
    .pipe(data(dataView))
    .pipe(pug())
    .pipe(rename({ extname: '.txt' }))
    .pipe(gulp.dest(`${env.outputFolder}`))
}

function mail () {
  return gulp.src('assets/files/mail.php', { encoding: false })
    .pipe(gulp.dest(`${env.outputFolder}`))
}

export default {
  build: gulp.series(htaccess, robots, mail)
}
