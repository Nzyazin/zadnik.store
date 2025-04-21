// yarn add -D webpack webpack-stream babel-loader @babel/core @babel/preset-env

import gulp from 'gulp'
import webpack from 'webpack-stream'

import env from './env.js'
import { browserSyncInstance } from './browserSync.js'

const path = {
  entry: 'assets/scripts/*.js',
  watch: 'assets/scripts/**/*.js'
}

const webpackConfig = function () {
  return {
    mode: env.production ? 'production' : 'development',
    entry: {
      app: './assets/scripts/app.js'
    },
    output: {
      filename: env.production ? `script-${env.hash}.js` : 'script.js',
    },
    devtool: env.production ? false : 'source-map',
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          use: {
            loader: 'babel-loader',
            options: {
              presets: [
                ['@babel/preset-env', {
                  modules: false,
                  useBuiltIns: 'entry',
                  corejs: { version: "3.9.1", proposals: true },
                }]
              ]
            }
          }
        }
      ]
    }
  }
}

const scripts = function () {
  return gulp.src(path.entry)
    .pipe(webpack(webpackConfig()))
    .pipe(gulp.dest(`${env.outputFolder}/statics/scripts`))
    .on('end', browserSyncInstance.reload)
}

export default {
  build: scripts,
  watchPath: path.watch
}
