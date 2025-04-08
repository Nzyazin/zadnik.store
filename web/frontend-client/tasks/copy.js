// yarn add -D gulp-plumber gulp-pug gulp-data gulp-rename

import gulp from 'gulp'
import plumber from 'gulp-plumber'
import pug from 'gulp-pug'
import data from "gulp-data"
import rename from 'gulp-rename'

import env from './env.js'

const dataView = { URL: env.url }

export default {
  build: gulp.series()
}
