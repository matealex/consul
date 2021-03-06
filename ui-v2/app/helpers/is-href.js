/*eslint ember/no-observers: "warn"*/
// TODO: Remove ^
import Helper from '@ember/component/helper';
import { inject as service } from '@ember/service';
import { observer } from '@ember/object';

export default Helper.extend({
  router: service('router'),
  compute(params) {
    return this.router.isActive(...params);
  },
  onURLChange: observer('router.currentURL', function() {
    this.recompute();
  }),
});
