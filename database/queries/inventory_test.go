package queries

// migratedQueries pins the complete bytes of the 21 SQL files copied from the
// legacy query client. These fixtures are independent of a sibling legacy checkout.
//
// Entries with fixed set were changed on purpose in Phase 4, and their pin is the
// fixed bytes. The legacy originals stay in infrastructure/legacy. Every other
// entry still pins the bytes copied at extraction time. INSERT_image.sql also
// moved, from events/images/confirm/ to images/create/.
var migratedQueries = []struct {
	path   string
	size   int
	sha256 string
	fixed  bool
}{
	{"authorization/clubs/can_manage/IS_student_authorized_club.sql", 418, "8e581ec057dcf332de2cceb54dbb845d68608f3f373f52d86da5588b246bb10b", false},
	{"authorization/events/can_manage/IS_student_authorized_event.sql", 539, "9b36b120e86d53ce1762a35f6af8a5279c8ff51eabce2933257ee34ac5060a17", false},
	{"clubs/create/INSERT_club.sql", 54, "a934bef1c2ffc2a55491bd1bd75f89606ab322c6bb1e64494aff076c2fc2a57d", false},
	{"clubs/create/INSERT_club_info.sql", 78, "050e1bb658dcc126054bbfa39d5929e9f48ed0b564128b970d216dd02db2601a", false},
	{"clubs/events/list/SELECT_club_events.sql", 1519, "d8aa2b947a877dc0244dd249245e1af4abb373d750b880f782eb3dd9116ee871", true},
	{"clubs/get/SELECT_club.sql", 539, "ef6ca31de0eca939b258e8c7abfad25d070dc3d0849b91a8b2c0dbf3c3fc747b", true},
	{"clubs/list/SELECT_clubs.sql", 590, "2385091f4bf6b36080db0484ae3e9316bcae3d575c1ecca71d4443a27eae3f9f", true},
	{"clubs/members/create/INSERT_club_member.sql", 626, "0503a101afe55ca32004daa4d90b52aaae07534e4469372a6bf5ff2bda16bbd1", false},
	{"clubs/members/leave/DELETE_club_member.sql", 197, "df9a34c77e5f023eb0819d3034726897bc2b8e99d50ce1aa6c657c1e13975b8b", false},
	{"events/create/INSERT_event.sql", 792, "8384888e2354e4aea211398588f8472c1d2d36118f1cae149d9da48012363bd3", true},
	{"events/create/INSERT_event_club_link.sql", 130, "2a039ddf9368d76e3eb0b70b862977542e266869f91f0ac2738c8042b2ff0591", false},
	{"events/create/INSERT_event_description.sql", 335, "742da857641ef7c307b5588fb56a3401180b6b2156059f35a29fbf4fe7bbedeb", true},
	{"events/images/confirm/INSERT_event_image.sql", 74, "7cfedca2b318385f7747228674b9828c4dbc985d7db445df7e0dea4c410245cd", false},
	{"images/create/INSERT_image.sql", 959, "99f3b079b861d64497d6814c51b4ab3956e6a93c8f3b58a10b50264d0cfa35fd", true},
	{"events/read/SELECT_events.sql", 1374, "ae5f42a9ed4c0d07a2309e8c3e2a12056eced4cb821785ed3e3e83507ab40fc0", true},
	{"me/clubs/list/SELECT_student_clubs.sql", 385, "bc6b4333e217b235d570097fed286bf7052256eefdcc65fcbc545602fc57d18f", false},
	{"me/events/list/SELECT_student_events.sql", 1533, "195d77d8b3cda8581a72b9415dbf0d36176672852031040d134e7524c776775a", true},
	{"me/get/SELECT_student_by_sub.sql", 189, "abe30d5eded05b5bea2162a1e93fcee66b20142bb47e5e0fac8b5833a96ddc47", false},
	{"students/ensure/EXISTS_student_by_sub.sql", 278, "1c36fc43596da415f30e3e15d88a2659fb02a22761209e88e0b4adc8e82a3a96", false},
	{"students/ensure/UPSERT_student.sql", 146, "b281775ddf5418ee7d6ed0b744b13740ba917b5072e84f6b01d95ad1462be9fb", false},
	{"students/ensure/UPSERT_student_sub_only.sql", 239, "1b5ec7e066fc3461115d756d2cf8d1c1cfe73e26b7a038574e7ff175b4b04d06", false},
}
