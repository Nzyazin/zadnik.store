import gulp from 'gulp'

import env from './env.js'

const path = 'assets/fonts/*.*'

function fonts () {
  return gulp.src(path, { encoding: false })
    .pipe(gulp.dest(`${env.outputFolder}/statics/fonts`))
}

export default {
  build: fonts
}
