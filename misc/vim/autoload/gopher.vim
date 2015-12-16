function! gopher#hey_gopher(...)
  let ret = system('gopherc -m ' . shellescape(join(a:000, ' ')))
  if v:shell_error != 0
    echohl Error | echom substitute(ret, '[\r\n]', '', 'g') | echohl None
  endif
endfunction

function! gopher#good_bye_gophers(...)
  let ret = system('gopherc -x ')
  if v:shell_error != 0
    echohl Error | echom substitute(ret, '[\r\n]', '', 'g') | echohl None
  endif
endfunction


