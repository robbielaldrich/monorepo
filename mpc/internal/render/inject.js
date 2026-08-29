// Injected into every Card Conjurer document *before* its own scripts run.
//
// Two jobs:
//   1. Count in-flight image loads so Go can tell when a redraw has settled.
//      Card Conjurer redraws from Image.onload callbacks and never reports
//      when it is done, so this is the only reliable "finished" signal.
//   2. Expose window.__mpc.render(spec), which drives the creator UI the same
//      way a person would and hands back the finished canvas.
(function () {
  if (window.__mpc) return;

  var pending = 0;
  var desc = Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, 'src');
  Object.defineProperty(HTMLImageElement.prototype, 'src', {
    configurable: true,
    enumerable: desc.enumerable,
    get: function () { return desc.get.call(this); },
    set: function (value) {
      var img = this;
      if (img.__mpcOpen) { img.__mpcOpen = false; pending--; }
      img.__mpcOpen = true;
      pending++;
      var done = function () {
        if (!img.__mpcOpen) return;
        img.__mpcOpen = false;
        pending--;
      };
      img.addEventListener('load', done, { once: true });
      img.addEventListener('error', done, { once: true });
      // A src that never resolves (dead URL, blocked host) must not wedge the
      // whole run, so every load gets a hard ceiling.
      setTimeout(done, 20000);
      desc.set.call(this, value);
    }
  });

  var mpc = window.__mpc = {
    pending: function () { return pending; },
    log: [],
    loadedPacks: {}
  };

  function sleep(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }

  // Resolves once no image is loading and nothing new starts for a beat.
  mpc.settle = async function (quietMs) {
    quietMs = quietMs || 120;
    var deadline = Date.now() + 30000;
    var quietSince = null;
    for (;;) {
      if (pending > 0) {
        quietSince = null;
      } else if (quietSince === null) {
        quietSince = Date.now();
      } else if (Date.now() - quietSince >= quietMs) {
        return;
      }
      if (Date.now() > deadline) throw new Error('timed out waiting for images');
      await sleep(30);
    }
  };

  // htmx swaps the creator fragment in and *then* appends its <script> tags,
  // so the DOM is ready well before creator-23.js has run. Waiting on an
  // element is not enough; wait for its globals. (Before the script runs,
  // `autoFrame` resolves to the <select id="autoFrame"> via named access,
  // which is why this checks for a function specifically.)
  mpc.waitForCreator = async function () {
    var deadline = Date.now() + 30000;
    for (;;) {
      if (typeof drawCard === 'function' &&
          typeof autoFrame === 'function' &&
          typeof loadFramePack === 'function' &&
          typeof resetCardIrregularities === 'function' &&
          typeof card === 'object' && card !== null) {
        return;
      }
      if (Date.now() > deadline) throw new Error('Card Conjurer did not finish loading');
      await sleep(25);
    }
  };

  mpc.loadPack = function (pack) {
    if (mpc.loadedPacks[pack]) return Promise.resolve();
    return new Promise(function (resolve, reject) {
      var s = document.createElement('script');
      s.src = '/js/frames/' + pack + '.js';
      s.onload = function () { mpc.loadedPacks[pack] = true; resolve(); };
      s.onerror = function () { reject(new Error('frame pack not found: ' + pack)); };
      document.head.appendChild(s);
    });
  };

  // Card Conjurer only draws with fonts the browser has already fetched.
  mpc.loadFonts = async function () {
    var faces = [];
    document.fonts.forEach(function (f) { faces.push(f); });
    await Promise.all(faces.map(function (f) { return f.load().catch(function () {}); }));
    await document.fonts.ready;
  };

  function el(sel) {
    var e = document.querySelector(sel);
    if (!e) throw new Error('missing element ' + sel);
    return e;
  }

  function setText(key, value) {
    if (value === undefined || value === null) return;
    if (!card.text || !card.text[key]) return;
    card.text[key].text = value;
  }

  mpc.render = async function (spec) {
    mpc.log = [];

    // 1. Frame version. Loading the pack defines availableFrames and wires up
    //    #loadFrameVersion, whose handler sets card.version, artBounds and the
    //    default text boxes for that frame style.
    await mpc.loadPack(spec.pack);
    var loadBtn = el('#loadFrameVersion');
    if (!loadBtn.disabled && loadBtn.onclick) {
      await loadBtn.onclick();
    }
    await mpc.settle();

    // 2. Card text.
    setText('title', spec.title);
    setText('mana', spec.mana);
    setText('type', spec.type);
    setText('rules', spec.rules);
    setText('pt', spec.pt);
    setText('nickname', spec.nickname);

    // 3. Collector info along the bottom.
    el('#info-number').value = spec.number || '';
    el('#info-rarity').value = spec.rarity || '';
    el('#info-set').value = spec.set || '';
    el('#info-language').value = spec.language || '';
    el('#info-note').value = spec.note || '';
    if (spec.year) el('#info-year').value = spec.year;
    el('#info-artist').value = spec.artist || '';
    el('#art-artist').value = spec.artist || '';
    el('#enableCollectorInfo').checked = spec.collector !== false;

    // 4. Frame stack. autoFrame() reads the text we just set and assembles the
    //    whole layer list, including legend crowns and hybrid half-frames.
    if (spec.frame) {
      el('#autoFrame').value = spec.frame;
      await autoFrame();
      await mpc.settle();
    }

    // Card Conjurer's own copyright line is skipped by the collector-info
    // renderer; giving it different text opts it back in.
    if (card.bottomInfo && card.bottomInfo.wizards) {
      if (spec.copyright === 'none' || spec.copyright === '') {
        card.bottomInfo.wizards.text = '';
      } else if (spec.copyright) {
        card.bottomInfo.wizards.text = '{ptshift0,0.0172}' + spec.copyright;
      }
    }

    // 5. Set symbol.
    if (spec.setSymbolUrl) {
      await mpc.setSymbol(spec.setSymbolUrl);
    } else if (spec.setSymbol) {
      el('#set-symbol-code').value = spec.setSymbol;
      el('#set-symbol-rarity').value = spec.rarity || 'c';
      fetchSetSymbol();
      await mpc.settle();
    } else {
      await mpc.setSymbol('/img/blank.png');
    }

    // 6. Watermark. The page is reused for every card in the deck, so an
    //    absent value has to actively clear what the last card left behind.
    if (spec.watermark) {
      await mpc.image(spec.watermark);
      uploadWatermark(spec.watermark);
      await mpc.settle();
      if (spec.watermarkOpacity) {
        el('#watermark-opacity').value = spec.watermarkOpacity * 100;
        watermarkEdited();
      }
    } else {
      uploadWatermark('/img/blank.png');
      await mpc.settle();
    }

    // 7. Art.
    if (spec.art) {
      await mpc.art(spec);
    } else {
      art.src = '/img/blank.png';
      el('#art-zoom').value = 100;
      el('#art-x').value = 0;
      el('#art-y').value = 0;
      el('#art-rotate').value = 0;
      await mpc.settle();
      artEdited();
    }

    // 8. Redraw every layer now that all inputs are final.
    await mpc.loadFonts();
    card.noCorners = false;
    drawTextBuffer();
    drawFrames();
    watermarkEdited();
    await bottomInfoEdited();
    await mpc.settle();
    drawCard();

    // A full-resolution PNG is several megabytes, which is far too much to
    // hand back through a CDP evaluate result. Post the bytes to the local
    // server instead and return only metadata.
    await mpc.upload('rounded');
    card.noCorners = true;
    drawCard();
    await mpc.upload('square');
    card.noCorners = false;
    drawCard();

    return {
      width: cardCanvas.width,
      height: cardCanvas.height,
      save: mpc.saveJSON(),
      warnings: mpc.log
    };
  };

  mpc.upload = function (name) {
    return new Promise(function (resolve, reject) {
      cardCanvas.toBlob(function (blob) {
        if (!blob) { reject(new Error('canvas produced no image')); return; }
        fetch('/_mpc/put/' + name, { method: 'POST', body: blob })
          .then(function (r) {
            if (!r.ok) throw new Error('upload failed: ' + r.status);
            resolve();
          })
          .catch(reject);
      }, 'image/png');
    });
  };

  mpc.image = function (src) {
    return new Promise(function (resolve, reject) {
      var i = new Image();
      i.crossOrigin = 'anonymous';
      i.onload = function () { resolve(i); };
      i.onerror = function () { reject(new Error('could not load image ' + src)); };
      i.src = src;
    });
  };

  mpc.setSymbol = async function (src) {
    await mpc.image(src);
    uploadSetSymbol(src, 'resetSetSymbol');
    await mpc.settle();
  };

  mpc.art = async function (spec) {
    var img = await mpc.image(spec.art);
    art.crossOrigin = 'anonymous';
    art.src = spec.art;
    await mpc.settle();

    var b = card.artBounds || { x: 0, y: 0, width: 1, height: 1 };
    var boxW = scaleWidth(b.width), boxH = scaleHeight(b.height);
    var fit = spec.artFit || 'cover';

    if (fit === 'manual') {
      el('#art-zoom').value = (spec.artZoom || 1) * 100;
      el('#art-x').value = Math.round((spec.artX || 0) * card.width);
      el('#art-y').value = Math.round((spec.artY || 0) * card.height);
    } else {
      // "cover" is autoFitArt's behaviour: match the tighter axis so the art
      // box is completely filled. "contain" matches the looser one instead.
      var wide = img.width / img.height > boxW / boxH;
      var zoom = (fit === 'contain') === wide ? boxW / img.width : boxH / img.height;
      el('#art-zoom').value = (zoom * 100).toFixed(2);
      el('#art-x').value = Math.round(scaleX(b.x) - (zoom * img.width - boxW) / 2 - scaleWidth(card.marginX));
      el('#art-y').value = Math.round(scaleY(b.y) - (zoom * img.height - boxH) / 2 - scaleHeight(card.marginY));
    }
    // Nudges are expressed as a fraction of the card so they survive a
    // change of frame or resolution.
    if (spec.artNudgeX) el('#art-x').value = Math.round(+el('#art-x').value + spec.artNudgeX * card.width);
    if (spec.artNudgeY) el('#art-y').value = Math.round(+el('#art-y').value + spec.artNudgeY * card.height);
    if (spec.artZoomScale && fit !== 'manual') {
      el('#art-zoom').value = (+el('#art-zoom').value * spec.artZoomScale).toFixed(2);
    }
    el('#art-rotate').value = spec.artRotate || 0;
    artEdited();
    await mpc.settle();
  };

  // A save file the real Card Conjurer UI can load, for hand tweaking.
  mpc.saveJSON = function () {
    var copy = JSON.parse(JSON.stringify(card, function (k, v) {
      return k === 'image' ? undefined : v;
    }));
    return copy;
  };
})();
