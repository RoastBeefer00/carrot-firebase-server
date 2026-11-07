// Page behavior that htmx and Alpine don't cover: dragging meals between days,
// keeping a list's scroll position when the page re-renders, and marking today.
// Everything listens on document, so it keeps working after #page is replaced.
(() => {
  // ---------- re-render helpers ----------

  // Lists that scroll on their own (the recipes sidebar) would jump back to the
  // top when #page is replaced, so their position is carried across the swap.
  const scrollPositions = {};
  document.addEventListener('htmx:beforeSwap', () => {
    document.querySelectorAll('[data-keep-scroll]').forEach((el) => {
      scrollPositions[el.dataset.keepScroll] = el.scrollTop;
    });
  });
  document.addEventListener('htmx:afterSwap', (e) => {
    document.querySelectorAll('[data-keep-scroll]').forEach((el) => {
      const top = scrollPositions[el.dataset.keepScroll];
      if (top) el.scrollTop = top;
    });
    // Opening a recipe from the list starts at its top.
    if (e.detail.requestConfig?.elt?.hasAttribute?.('data-scroll-top')) window.scrollTo(0, 0);
    markToday();
  });

  function markToday() {
    const today = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][new Date().getDay()];
    document.querySelectorAll('[data-weekday]').forEach((el) => {
      el.classList.toggle('is-today', el.dataset.weekday === today);
    });
  }
  document.addEventListener('DOMContentLoaded', markToday);

  // ---------- drag and drop ----------
  // One pointer-based path for mouse and touch, since phones don't fire HTML5
  // drag events. The grip picks a meal up at once. Elsewhere on a meal, a mouse
  // picks it up as soon as it moves and a finger holds for a moment first, so a
  // plain swipe still scrolls. Dropping on a meal on another day swaps their
  // days; dropping anywhere else in a day moves the meal there.
  let drag = null;
  let suppressClick = false;

  function dropTargetAt(x, y) {
    const hit = document.elementFromPoint(x, y);
    if (!hit) return null;
    const meal = hit.closest('[data-drag]');
    if (meal && meal.dataset.drag !== drag.id) {
      const zone = meal.closest('[data-drop]');
      if (zone && zone.dataset.drop !== drag.fromDay) return { el: meal, swap: meal.dataset.drag };
    }
    const zone = hit.closest('[data-drop]');
    return zone ? { el: zone, day: zone.dataset.drop } : null;
  }

  function highlightDrop() {
    const t = dropTargetAt(drag.x, drag.y);
    document.querySelectorAll('.drop, .swap').forEach((el) => {
      if (el !== t?.el) el.classList.remove('drop', 'swap');
    });
    t?.el.classList.add(t.swap ? 'swap' : 'drop');
  }

  function moveGhost() {
    drag.ghost.style.left = `${drag.x - drag.dx}px`;
    drag.ghost.style.top = `${drag.y - drag.dy}px`;
  }

  // Scrolls the page, or the list being dragged in, while the pointer is near
  // its top or bottom edge.
  function autoScroll() {
    if (!drag) return;
    const s = drag.scroller;
    const top = s ? s.getBoundingClientRect().top : 56;
    const bottom = s ? s.getBoundingClientRect().bottom : window.innerHeight;
    const edge = 70;
    let dy = 0;
    if (drag.y < top + edge) dy = -Math.ceil((top + edge - drag.y) / 6);
    else if (drag.y > bottom - edge) dy = Math.ceil((drag.y - (bottom - edge)) / 6);
    if (dy) {
      s ? (s.scrollTop += dy) : window.scrollBy(0, dy);
      highlightDrop();
    }
    requestAnimationFrame(autoScroll);
  }

  function scrollerFor(el) {
    for (let p = el.parentElement; p && p !== document.body; p = p.parentElement) {
      const o = getComputedStyle(p).overflowY;
      if ((o === 'auto' || o === 'scroll') && p.scrollHeight > p.clientHeight) return p;
    }
    return null;
  }

  function startDrag(el, x, y) {
    const rect = el.getBoundingClientRect();
    const ghost = el.cloneNode(true);
    ghost.classList.add('ghost');
    ghost.style.width = `${rect.width}px`;
    document.body.appendChild(ghost);
    el.classList.add('lifted');
    document.body.classList.add('dragging');
    navigator.vibrate?.(15);
    drag = {
      id: el.dataset.drag,
      fromDay: el.closest('[data-drop]')?.dataset.drop,
      el,
      ghost,
      x,
      y,
      dx: x - rect.left,
      dy: y - rect.top,
      scroller: scrollerFor(el),
    };
    suppressClick = true;
    moveGhost();
    highlightDrop();
    requestAnimationFrame(autoScroll);
  }

  // The page is re-rendered from the server after a drop, but the meal moves
  // right away so the drop doesn't wait on the request.
  function moveInPlace(el, zone) {
    const items = zone.querySelector('[data-items]');
    if (!items) return;
    const from = el.closest('[data-drop]');
    const node = el.closest('li') ?? el;
    items.insertBefore(node, items.querySelector('[data-empty]'));
    zone.querySelector('[data-empty]')?.classList.add('hidden');
    if (from && !from.querySelector('[data-drag]')) from.querySelector('[data-empty]')?.classList.remove('hidden');
  }

  function swapInPlace(a, b) {
    const na = a.closest('li') ?? a;
    const nb = b.closest('li') ?? b;
    const mark = document.createComment('');
    na.replaceWith(mark);
    nb.replaceWith(na);
    mark.replaceWith(nb);
  }

  function send(path, values) {
    htmx.ajax('POST', path, { source: document.getElementById('page'), target: '#page', swap: 'innerHTML', values });
  }

  function endDrag(commit) {
    if (!drag) return;
    const { id, el, ghost, fromDay } = drag;
    const target = commit ? dropTargetAt(drag.x, drag.y) : null;
    ghost.remove();
    el.classList.remove('lifted');
    document.body.classList.remove('dragging');
    document.querySelectorAll('.drop, .swap').forEach((x) => x.classList.remove('drop', 'swap'));
    drag = null;
    setTimeout(() => {
      suppressClick = false;
    }, 100);
    if (!target) return;
    if (target.swap) {
      swapInPlace(el, target.el);
      send(`/api/plan/swap/${id}`, { with: target.swap });
    } else if (target.day !== fromDay) {
      moveInPlace(el, target.el);
      send(`/api/plan/day/${id}`, { day: target.day });
    }
  }

  document.addEventListener('pointerdown', (e) => {
    const el = e.target.closest('[data-drag]');
    if (!el || drag || e.button !== 0 || e.target.closest('button, .plan-quick')) return;
    if (e.target.closest('[data-grip]')) {
      e.preventDefault();
      startDrag(el, e.clientX, e.clientY);
      return;
    }
    const mouse = e.pointerType === 'mouse';
    const sx = e.clientX;
    const sy = e.clientY;
    let lx = sx;
    let ly = sy;
    const timer = mouse ? null : setTimeout(() => {
      cleanup();
      startDrag(el, lx, ly);
    }, 320);
    const move = (ev) => {
      lx = ev.clientX;
      ly = ev.clientY;
      if (Math.hypot(lx - sx, ly - sy) <= (mouse ? 4 : 8)) return;
      clearTimeout(timer);
      cleanup();
      if (mouse) {
        startDrag(el, sx, sy);
        drag.x = lx;
        drag.y = ly;
        moveGhost();
        highlightDrop();
      }
    };
    const stop = () => {
      clearTimeout(timer);
      cleanup();
    };
    function cleanup() {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', stop);
      window.removeEventListener('pointercancel', stop);
    }
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', stop);
    window.addEventListener('pointercancel', stop);
  });

  // A drag ends in a click on the meal, which would open it.
  document.addEventListener('click', (e) => {
    if (suppressClick) {
      e.stopPropagation();
      e.preventDefault();
      suppressClick = false;
    }
  }, true);
  // Links are draggable by default, which would start the browser's own drag.
  document.addEventListener('dragstart', (e) => {
    if (e.target.closest?.('[data-drag]')) e.preventDefault();
  });
  // Stops the long-press menu on touch screens.
  document.addEventListener('contextmenu', (e) => {
    if (e.target.closest('[data-drag]') && e.pointerType !== 'mouse') e.preventDefault();
  });
  // Once a drag starts, the finger moves the meal instead of scrolling the page.
  window.addEventListener('touchmove', (e) => {
    if (drag) e.preventDefault();
  }, { passive: false });
  window.addEventListener('pointermove', (e) => {
    if (!drag) return;
    drag.x = e.clientX;
    drag.y = e.clientY;
    moveGhost();
    highlightDrop();
  });
  window.addEventListener('pointerup', () => endDrag(true));
  window.addEventListener('pointercancel', () => endDrag(false));
})();
