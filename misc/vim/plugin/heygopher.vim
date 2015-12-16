if !executable('gopherc')
  finish
endif

command! -nargs=+ HeyGopher call gopher#hey_gopher(<f-args>)
command! -nargs=0 GoodByeGophers call gopher#good_bye_gophers(<f-args>)
