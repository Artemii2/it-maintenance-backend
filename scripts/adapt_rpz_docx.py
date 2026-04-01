# -*- coding: utf-8 -*-
"""
Адаптация РПЗ (Word): замена темы «подразделения» на вариант «тормозные колодки / IT-maintenance».
Исходный распакованный docx: d:\\IU5-65B_Berdnikov_RPZ_unpack
Результат: d:\\IU5-65B_Berdnikov_RPZ_adapted.docx
Картинки и разметка не трогаются — меняется только текст в word/document.xml.
"""

from __future__ import annotations

import shutil
import zipfile
from pathlib import Path

SRC_UNPACK = Path(r"d:\IU5-65B_Berdnikov_RPZ_unpack")
OUT_DOCX = Path(r"d:\IU5-65B_Berdnikov_RPZ_adapted.docx")
WORK = Path(r"d:\IU5-65B_Berdnikov_RPZ_work")
XML_REL = Path("word/document.xml")

# Порядок: сначала длинные уникальные подстроки, затем короткие.
REPLACEMENTS: list[tuple[str, str]] = [
    # --- Тема, цель, назначение, нагрузка ---
    (
        "Административная структура компании",
        "Веб-сервис расчёта износа тормозных колодок (вариант IT-maintenance)",
    ),
    (
        "Необходимо разработать систему, включающую в себя вебсервис, вебприложение и нативное приложение, которая позволит сотрудникам формировать и отправлять заявки по подразделениям организации.",
        "Необходимо разработать систему, включающую в себя веб-сервис, веб-приложение (SPA) и нативное приложение, которая позволит пользователям формировать и отправлять заявки на расчёт остаточного ресурса тормозных колодок.",
    ),
    (
        "Необходимо разработать систему, которая будет автоматизировать работу с заявками по подразделениям и предоставлять актуальную информацию о них.",
        "Необходимо разработать систему, которая будет автоматизировать работу с заявками на расчёт износа тормозных колодок и предоставлять актуальную информацию о прогнозе остаточного ресурса.",
    ),
    (
        " смогут создавать заявки, выбирая подразделения, указывая их роли в заявке и связанные показатели (например, заработную плату) при заполнении формы. ",
        " смогут создавать заявки, выбирая услуги каталога (комплекты тормозных колодок), указывая количество и комментарии к позициям при заполнении формы. ",
    ),
    (
        " смогут принять заявку к рассмотрению, утвердить её либо отклонить. ",
        " смогут завершить обработку заявки (подтвердить расчёт) либо отклонить заявку. ",
    ),
    (
        " будет доступен список их заявок, включающий активные и обработанные. Гости системы имеют доступ к описанию подразделений и могут пройти регистрацию для последующей работы с заявками.",
        " будет доступен список их заявок, включающий активные и обработанные. Гости системы имеют доступ к каталогу услуг (тормозные колодки) и могут пройти регистрацию для последующей работы с заявками.",
    ),
    (
        "по подразделениям. Система рассчитана на одновременную работу до 100 сотрудников и 10 руководителей.",
        "по расчёту износа. Система рассчитана на одновременную работу до 100 пользователей-создателей заявок и 10 модераторов.",
    ),
    (
        "формирует 30 заявок по подразделениям в день, его рабочий день длится 8 часов.",
        "формирует 30 заявок на расчёт износа в день, его рабочий день длится 8 часов.",
    ),
    # --- Вводный абзац про таблицы (с неразрывным пробелом перед скобкой в конце фрагмента) ---
    (
        " (таблица 1). Для хранения состава одной заявки по подразделениям используется таблица ",
        " (таблица 1). Для связи заявки с выбранными услугами каталога (связь «многие-ко-многим») используется таблица ",
    ),
    (
        " (таблица 3) представляет собой список всех заявок по подразделениям. Данные о ",
        " (таблица 3) представляет собой список заявок на расчёт износа тормозных колодок. Данные о ",
    ),
    (
        "Данные о подразделениях хранятся в таблице ",
        "Каталог тормозных колодок (услуги) хранится в таблице ",
    ),
    (
        "сотрудниках и руководителях (модераторах заявок) хранятся в таблице users",
        "пользователях и модераторах заявок хранятся в таблице users",
    ),
    # --- Размеры записей (формулы 5–6) ---
    (
        " заявки и двух пользователей – по 4 байта, три системные даты – по 8 байт, статус – 20 байт, заголовок заявки – 255 байт) и ",
        " заявки (id) и внешние ключи пользователей (created_by_id, moderator_id) — по 8 байт, даты created_at/formed_at/completed_at — по 8 байт, статус — 16 байт, стиль вождения — 64 байта, пробег mileage — 4 байта) и ",
    ),
    (
        " (четыре целочисленных поля – по 4 байта, роль подразделения – 100 байт, значение заработной платы – 16 байт) по формулам 5 и 6.",
        " (ключи application_id и service_id — по 8 байт, quantity и position — по 4 байта, is_primary — 1 байт; remaining_km — 8 байт, remaining_percent — 4 байта) по формулам 5 и 6.",
    ),
    (
        "на заявки по подразделениям.",
        "на заявки по расчёту износа колодок.",
    ),
    # --- Аппаратные требования: абзац разбит на два <w:t> (A заканчивается на «…заявок », B — «по подразделениям. Система…») ---
    (
        # После «заявок» в XML — обычный пробел перед </w:t>, не NBSP (иначе replace не срабатывает).
        "а основе анализа структурных данных и предполагаемой нагрузки был произведён расчёт аппаратных требований для системы заявок ",
        "На основе анализа структурных данных и предполагаемой нагрузки был произведён расчёт аппаратных требований для веб-сервиса заявок на расчёт износа тормозных колодок. ",
    ),
    (
        "по расчёту износа. Система рассчитана",
        "Система рассчитана",
    ),
    (
        "каждый руководитель  генерирует ещё по одному RPS",
        "каждый модератор генерирует ещё по одному RPS",
    ),
    # --- Этапы разработки (ещё фрагменты старого варианта) ---
    (
        "Создание базы данных для хранения информации об устройствах и заявках на расчёт нагрузки (Postgres, gorm)",
        "Создание базы данных для хранения каталога тормозных колодок, заявок и связи заявок с услугами (PostgreSQL, GORM)",
    ),
    (
        "Создание дизайна приложения в figma на основе дизайна 1c.ru, развёртывание Minio",
        "Создание дизайна интерфейса в Figma, развёртывание MinIO для медиафайлов",
    ),
    # «об устройствах» разбито Word на два run: «…о» + «б устройствах»
    ("б устройствах", " каталоге тормозных колодок"),
    ("расчёт нагрузки", "расчёт износа (остаточный ресурс)"),
    # --- Грамматика после замен «подразделение» → «услуга» ---
    ("Получить одно услугу", "Получить одну услугу"),
    ("Сформировать заявку по заявкам на расчёт", "Сформировать заявку (перевод черновика в статус formed)"),
    # --- Приложение Б / таблицы: имена полей и DTO ---
    ("forming_date", "formed_at"),
    ("finish_date", "completed_at"),
    ("incomplete_items_count", "items_count"),
    ("employee_count", "base_resource"),
    ("is_deleted", "status"),
    ("photo_url", "image_url"),
    ("reports_to", "driving_style_hint"),
    ("short_description", "price"),
    ("<w:t>head</w:t>", "<w:t>pad_type</w:t>"),
    ("head: string, ", "pad_type: string, "),
    ("salary: float|null", "quantity: int"),
    ("salary: float", "quantity: int"),
    ("<w:t>salary</w:t>", "<w:t>quantity</w:t>"),
    ('<w:t xml:space="preserve">salary: </w:t>', '<w:t xml:space="preserve">quantity: </w:t>'),
    (": int, role: string, salary: ", ": int, quantity: int, comment: string, "),
    ("applications_service", "application_services"),
    # --- Имена таблиц (англ.) ---
    ("department_application_departments", "application_services"),
    ("department_applications_department", "application_services"),
    ("department_applications", "applications"),
    ("departments_count", "items_count"),
    ("dep_app_dep", "brake-pad-wear"),
    ("main_department_id", "service_id"),
    # --- HTTP-пути: в Word URL разбиты на несколько <w:t>, поэтому без лишнего «/api» внутри куска ---
    ("/users/signup", "/users/register"),
    ("<w:t>/users/</w:t>", "<w:t>/auth/</w:t>"),
    ("<w:t>signin</w:t>", "<w:t>login</w:t>"),
    ("<w:t>signout</w:t>", "<w:t>logout</w:t>"),
    ("/departments", "/brake-pad"),
    ("/department/{id}", "/brake-pad/{id}"),
    ("/department/create-department", "/brake-pad"),
    # корзина: было department_application + -cart
    ("<w:t>-cart</w:t>", "<w:t>/cart-icon</w:t>"),
    # список заявок: убрать /all- + applications (см. apply_post_fixes)
    # JSON / описания полей в приложении Б
    ("{ department_application_id:", "{ id:"),
    ("department_application_id: int", "id: int"),
    ("department_id: int", "service_id: int"),
    ("signup", "register"),
    # не заменять подстроку signin→login: испортит уже подставленный /api/auth/login
    # --- Роли и интерфейс (общие формулировки) ---
    ("Доступно аутентифицированному администратору", "Доступно аутентифицированному пользователю (создателю заявки или модератору)"),
    (", доступно только администратору", ", доступно только модератору"),
    ("Сотрудникам будет", "Пользователям будет"),
    ("Руководитель смогут", "Модератор сможет"),
    ("если он не администратор", "если он не модератор"),
    ("Все поданные заявки ", "Все заявки в системе "),
    ("административное объединение", "черновую заявку на расчёт износа"),
    ("Наименование должности/лица, которому подчиняется руководитель", "Рекомендуемый стиль вождения для услуги (подсказка)"),
    ("Руководитель подразделения", "Тип колодок (pad_type)"),
    ("Количество сотрудников в подразделении", "Базовый ресурс колодок, км (base_resource)"),
    ("Количество сотрудников", "Базовый ресурс, км"),
    ("Полное текстовое описание подразделения", "Полное текстовое описание услуги"),
    ("Флаг «мягкого удаления» подразделения", "Статус записи в каталоге (active/deleted)"),
    ("Ссылка на изображение подразделения", "Имя файла изображения (image_url)"),
    ("Ссылка на видео о подразделении", "Имя файла видео (video_url)"),
    ("Краткое описание подразделения", "Цена комплекта, руб."),
    ("Идентификатор основного подразделения", "Идентификатор услуги (service_id)"),
    ("Роль подразделения в заявке", "Порядок отображения позиции в заявке"),
    ("Значение заработной платы", "Количество комплектов (quantity)"),
    ("Порядок отображения подразделения в рамках заявки", "Позиция в списке позиций (position)"),
    ("Статус заявки по подразделениям", "Статус заявки на расчёт износа"),
    ("Первичный ключ, идентификатор заявки по подразделениям", "Первичный ключ заявки (id)"),
    ("Идентификатор создателя заявки", "Идентификатор создателя (created_by_id)"),
    ("creator_id", "created_by_id"),
    ("Заголовок заявки-черновика", "Параметры расчёта (пробег, стиль вождения)"),
    ("Заголовок заявки", "Расчётные поля заявки (в т.ч. min_remaining_km, min_remaining_percent)"),
    ("Дата и время формирования заявки", "Дата и время formed_at (оформление заявки)"),
    ('<w:t xml:space="preserve">руководитель </w:t>', '<w:t xml:space="preserve">модератор </w:t>'),
    ("Пароль пользователя", "Хэш пароля (password_hash)"),
    ("Флаг «является ли пользователь инженером»", "Флаг «является ли пользователь модератором»"),
    # --- Страницы и действия ---
    ("Список подразделений", "Каталог услуг (тормозные колодки)"),
    ("Одно подразделение", "Одна услуга каталога"),
    ("Получить список подразделений", "Получить список услуг каталога"),
    ("Получить одно подразделение", "Получить одну услугу"),
    ("Создать подразделение", "Создать услугу в каталоге"),
    ("Добавить подразделение в заявку", "Добавить услугу в черновик заявки"),
    ("Добавить – добавляет подразделение", "Добавить – добавляет услугу"),
    ("корзину заявки по подразделениям", "иконку корзины / черновик заявки"),
    ("Получить корзину заявки по подразделениям", "Получить сведения о черновике (корзине)"),
    ("Получить список всех заявок по подразделениям", "Получить список заявок на расчёт износа"),
    ("Получить одну заявку по подразделениям", "Получить одну заявку на расчёт износа"),
    ("с перечнем подразделений", "с плоским перечнем позиций (услуг)"),
    ("Изменить поля заявки по подразделениям", "Изменить поля заявки (пробег, стиль вождения, статус)"),
    ("Сформировать заявку по подразделениям", "Сформировать заявку (перевести черновик в статус formed)"),
    ("Завершить заявку по подразделениям", "Завершить заявку модератором (completed)"),
    ("Удалить подразделение из заявки", "Удалить позицию услуги из заявки"),
    ("Удалить (логически) заявку по подразделениям", "Удалить (логически) заявку"),
    (
        "Обновить данные о подразделении в заявке по подразделениям (роль, зарплата, порядок; либо сдвиг вверх/вниз). ",
        "Обновить позицию услуги в заявке (количество, комментарий, порядок; либо сдвиг вверх/вниз). ",
    ),
    ("Обновить данные о подразделении в заявке по подразделениям", "Обновить позицию услуги в заявке"),
    (" Обновить данные о подразделении", " Обновить позицию в заявке"),
    ("Оформить – отправляет текущую заявку-черновик ", "Оформить – переводит черновик в статус formed "),
    ("по подразделениям ", "на расчёт износа "),
    ("по подразделениям ", "на расчёт износа "),
    (" по подразделениям", " на расчёт износа"),
    ("по подразделениям", "по расчёту износа"),
    ("подразделениям", "заявкам на расчёт"),
    ("подразделения", "услуги каталога"),
    ("подразделений", "услуг"),
    ("подразделение", "услугу"),
    ("Отображается подробная информация выбранного ", "Отображается карточка выбранной "),
    ("Отображает текущую заявку-черновик пользователя ", "Отображает черновик заявки пользователя "),
    ("элементы карточек с устройствами", "карточки услуг каталога"),
    ("Удалить заявку – удаляет заявку-черновик на расчёт, (вызывается метод 4.1.1", "Удалить заявку – удаляет заявку-черновик (вызывается метод 4.1.14"),
    ("руководителей", "модераторов"),
    ("сотрудниках", "пользователях"),
    ("сотрудников", "пользователей"),
    ("сотрудникам", "пользователям"),
    (" сотрудник ", " пользователь "),
    ("Сотрудники", "Пользователи"),
    ("Сотрудникам", "Пользователям"),
    ("руководителях", "модераторах"),
    # --- Оставшиеся department* (после длинных замен) ---
    ("department_application_id", "application_id"),
    ("department_id", "service_id"),
    ("department_application", "brake-pad-wear"),
    ("departments", "services"),
    ("department", "service"),
    # users: колонки (после department*, чтобы не задеть чужие строки)
    ("user_id", "id"),
    ("{login: string, password: string, is_moderator: boolean}", "{username: string, password: string, is_moderator: boolean}"),
    ("{login: string, password: string }", "{username: string, password: string }"),
    ("{login: string, ", "{username: string, "),
    ("creator_login", "creator_username"),
    ("moderator_login", "moderator_username"),
]


def apply_replacements(text: str) -> str:
    # В документе ровно одно поле таблицы users «login» (раньше в файле, чем URL в приложении Б).
    # Сначала переименуем его в username, затем signin→login для пути /api/auth/login.
    text = text.replace("<w:t>login</w:t>", "<w:t>username</w:t>", 1)
    for old, new in REPLACEMENTS:
        text = text.replace(old, new)
    return text


# Фрагменты Word XML. После замены department_application → brake-pad-wear в приложении Б хвост URL такой:
#   /{id}/edit-  +  второй run «brake-pad-wear» — лишний, удаляем целиком (со spellStart/spellEnd).
_EDIT_TAIL = (
    '<w:t>/{id}/edit-</w:t></w:r><w:proofErr w:type="spellStart"/>'
    '<w:r w:rsidRPr="00BD5E5E"><w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" '
    'w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:sz w:val="28"/><w:szCs w:val="28"/>'
    '<w:lang w:val="en-US"/></w:rPr><w:t>brake-pad-wear</w:t></w:r><w:proofErr w:type="spellEnd"/>'
)
_FORM_TAIL = _EDIT_TAIL.replace("/{id}/edit-", "/{id}/form-")
_DELETE_TAIL = _EDIT_TAIL.replace("/{id}/edit-", "/{id}/delete-")
_FINISH_TAIL = (
    '<w:t>/{id}/finish-</w:t></w:r><w:proofErr w:type="spellStart"/>'
    '<w:r w:rsidRPr="00BD5E5E"><w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" '
    'w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:sz w:val="28"/><w:szCs w:val="28"/>'
    '<w:lang w:val="en-US"/></w:rPr><w:t>brake-pad-wear</w:t></w:r><w:proofErr w:type="spellEnd"/>'
)

_OLD_ADD_CART = (
    '<w:t>/add/{</w:t></w:r><w:proofErr w:type="spellStart"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>service_id</w:t></w:r><w:proofErr w:type="spellEnd"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>}</w:t></w:r>'
)

_OLD_ITEMS_PAIR = (
    '<w:t>/{</w:t></w:r><w:proofErr w:type="spellStart"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>service_id</w:t></w:r><w:proofErr w:type="spellEnd"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>}/{</w:t></w:r><w:proofErr w:type="spellStart"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>application_id</w:t></w:r><w:proofErr w:type="spellEnd"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>}</w:t></w:r>'
)

# Включать открывающий <w:r>…</w:rPr>, иначе после удаления остаётся «пустой» run без <w:t>.
_OLD_ALL_APPS = (
    '<w:r w:rsidRPr="00BD5E5E"><w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" '
    'w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>/all-</w:t></w:r><w:proofErr w:type="spellStart"/><w:r w:rsidRPr="00BD5E5E">'
    '<w:rPr><w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:t>applications</w:t></w:r><w:proofErr w:type="spellEnd"/>'
)


# Word иногда разбивает длинный токен после массовых замен (остаток «…_c» + «ount»).
_BROKEN_ITEMS_COUNT = (
    '<w:t>incomplete_items_c</w:t></w:r><w:r w:rsidRPr="00BD5E5E"><w:rPr>'
    '<w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>'
    '<w:sz w:val="28"/><w:szCs w:val="28"/><w:lang w:val="en-US"/></w:rPr>'
    '<w:lastRenderedPageBreak/><w:t>ount</w:t></w:r>'
)


def apply_post_fixes(text: str) -> str:
    _end = '<w:t>{inner}</w:t></w:r><w:proofErr w:type="spellEnd"/>'
    text = text.replace(_EDIT_TAIL, _end.format(inner="/{id}"))
    text = text.replace(_FORM_TAIL, _end.format(inner="/{id}"))
    text = text.replace(_DELETE_TAIL, _end.format(inner="/{id}"))
    text = text.replace(_FINISH_TAIL, _end.format(inner="/{id}/finish"))
    # список заявок: убрать /all-applications
    text = text.replace(_OLD_ALL_APPS, "")
    # POST добавления в корзину
    text = text.replace(_OLD_ADD_CART, '<w:t>/cart/items</w:t></w:r>')
    # DELETE/PUT позиции в заявке
    text = text.replace(_OLD_ITEMS_PAIR, '<w:t>/{id}/items/{serviceId}</w:t></w:r>')
    # не дублировать префикс api (в ячейке уже есть отдельный w:t «api»)
    text = text.replace("<w:t>/api/brake-pad</w:t>", "<w:t>/brake-pad</w:t>")
    text = text.replace(_BROKEN_ITEMS_COUNT, '<w:t>items_count</w:t></w:r>')
    return text


def main() -> None:
    if not SRC_UNPACK.is_dir():
        raise SystemExit(f"Нет папки распаковки: {SRC_UNPACK}")

    if WORK.exists():
        shutil.rmtree(WORK)
    shutil.copytree(SRC_UNPACK, WORK)

    xml_path = WORK / XML_REL
    if not xml_path.is_file():
        raise SystemExit(f"Не найден {xml_path}")

    text = xml_path.read_text(encoding="utf-8")
    text = apply_replacements(text)
    text = apply_post_fixes(text)
    xml_path.write_text(text, encoding="utf-8")

    out_path = OUT_DOCX
    if out_path.exists():
        try:
            out_path.unlink()
        except PermissionError:
            base = out_path.stem
            parent = out_path.parent
            for n in range(1, 50):
                candidate = parent / f"{base}_new{n}.docx"
                if not candidate.exists():
                    out_path = candidate
                    break
            else:
                raise SystemExit(f"Не удалось удалить и не найдено свободного имени рядом с {OUT_DOCX}")

    try:
        zf_ctx = zipfile.ZipFile(out_path, "w", compression=zipfile.ZIP_DEFLATED)
    except PermissionError:
        base = out_path.stem.rsplit("_new", 1)[0] if "_new" in out_path.stem else out_path.stem
        parent = out_path.parent
        for n in range(1, 50):
            candidate = parent / f"{base}_new{n}.docx"
            if candidate.exists():
                continue
            out_path = candidate
            break
        else:
            raise SystemExit(f"Нет прав на запись в {OUT_DOCX} и рядом")
        zf_ctx = zipfile.ZipFile(out_path, "w", compression=zipfile.ZIP_DEFLATED)

    with zf_ctx as zf:
        for path in WORK.rglob("*"):
            if path.is_file():
                arc = path.relative_to(WORK).as_posix()
                zf.write(path, arcname=arc)

    print(f"OK: {out_path}")


if __name__ == "__main__":
    main()
