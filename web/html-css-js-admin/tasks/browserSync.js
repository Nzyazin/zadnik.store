// yarn add -D browser-sync

import browserSync from 'browser-sync'

export const browserSyncInstance = browserSync.create()

export function runServe () {
  browserSyncInstance.init({
    server: {
      baseDir: 'dev',
      index: "_site-map.html"
    }
  })
}
