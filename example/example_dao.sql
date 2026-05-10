-- name: select_item_by_id
/* select_item_by_id : id로 단건 조회 (없으면 sql.ErrNoRows) */
select
	id
	,name
	,reg_dttm
from item
where id = :id
	and member_id = :member_id

-- name: insert_item_returning_id
/* insert_item_returning_id : 신규 INSERT 후 자동 부여된 id 반환 */
insert into item
(
	member_id
	,name
	,reg_pgm
)
values
(
	:member_id
	,:name
	,:reg_pgm
)
returning id

-- name: select_items_by_member
/* select_items_by_member : 회원의 모든 아이템 (생성 역순) */
select
	id
	,name
	,reg_dttm
from item
where member_id = :member_id
order by id desc
