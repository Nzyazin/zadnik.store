// yarn add -D del

import { deleteAsync } from 'del'

import env  from './env.js'

function cleanAll () {
  return deleteAsync([`${env.outputFolder}/**`])
}

export default {
  all: cleanAll,
}
