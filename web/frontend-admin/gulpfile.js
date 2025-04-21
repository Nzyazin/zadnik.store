// yarn add cross-env
// yarn add -D gulp

'use strict'

import gulp from 'gulp'
import { runServe } from './tasks/browserSync.js'
import clean from './tasks/clean.js'
import fonts from './tasks/fonts.js'
import img from './tasks/images.js'
import scripts from './tasks/scripts.js'
import styles from './tasks/styles.js'
import view from './tasks/view.js'

gulp.task('watch', () => {
  gulp.watch(view.watchPath, gulp.series(view.build))
  gulp.watch(styles.watchPath, gulp.series(styles.build))
  gulp.watch(scripts.watchPath, gulp.series(scripts.build))
  gulp.watch(img.watchPath, gulp.series(img.build))
})

gulp.task('dev', gulp.series(
  clean.all,
  fonts.build,
  styles.build,
  scripts.build,
  img.build,
  view.build
))

gulp.task('build', gulp.series(
  clean.all,
  fonts.build,
  styles.build,
  scripts.build,
  img.build,
  view.build
))

gulp.task('default', gulp.series(
  'dev',
  gulp.parallel(
    'watch', runServe
  )
))
