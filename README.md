**Осу парсер мп-линков для квалификаций и т.д.**

1. Для использования, необходимо зарегистрировать новое OAuth приложение в осу - https://osu.ppy.sh/home/account/edit#new-oauth-application
*(подробнее об этом написано тут - https://osu.ppy.sh/docs/index.html#registering-an-oauth-application)*
2. После чего указать Client Secret и Client ID в файл secrets.json (только не коммите этот файл с встроенными значениями на гитхаб!)


**TODO:**

1. ~~Определять дублирование ID карты и засчитывать только последний скор~~ Готово - нужно лишь потестировать этот функционал
2. ~~Сделать так, чтобы показывалось больше, чем 100 events, т.е. в случае долгих мп-линков (что-то с get-параметром before нужно думать)~~ Готово, но опять же нужно тестить и в случае чего - вносить правки
