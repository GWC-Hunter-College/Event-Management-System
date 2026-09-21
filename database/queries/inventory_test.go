package queries

// migratedQueries pins the complete bytes of the 21 legacy application SQL files
// at extraction time. These fixtures are independent of a sibling legacy checkout.
var migratedQueries = []struct {
	path   string
	size   int
	sha256 string
}{
	{"authorization/clubs/can_manage/IS_student_authorized_club.sql", 418, "8e581ec057dcf332de2cceb54dbb845d68608f3f373f52d86da5588b246bb10b"},
	{"authorization/events/can_manage/IS_student_authorized_event.sql", 539, "9b36b120e86d53ce1762a35f6af8a5279c8ff51eabce2933257ee34ac5060a17"},
	{"clubs/create/INSERT_club.sql", 54, "a934bef1c2ffc2a55491bd1bd75f89606ab322c6bb1e64494aff076c2fc2a57d"},
	{"clubs/create/INSERT_club_info.sql", 78, "050e1bb658dcc126054bbfa39d5929e9f48ed0b564128b970d216dd02db2601a"},
	{"clubs/events/list/SELECT_club_events.sql", 792, "7e75ddcec2cbebc31cb2379ff457b6ce836b11925933cf2a2b348f82815bcebd"},
	{"clubs/get/SELECT_club.sql", 282, "d611db5c7f33df2e408a66c3794ef500ccb61df180b32a505fb7569319619a8b"},
	{"clubs/list/SELECT_clubs.sql", 334, "0c5c4241fc435658203f5338d233cdc490fe6d471707e48d11e217e6acc1080e"},
	{"clubs/members/create/INSERT_club_member.sql", 626, "0503a101afe55ca32004daa4d90b52aaae07534e4469372a6bf5ff2bda16bbd1"},
	{"clubs/members/leave/DELETE_club_member.sql", 197, "df9a34c77e5f023eb0819d3034726897bc2b8e99d50ce1aa6c657c1e13975b8b"},
	{"events/create/INSERT_event.sql", 312, "ca2342754680a425f9379fd745876e52d74956a99422a46528581c134780a281"},
	{"events/create/INSERT_event_club_link.sql", 130, "2a039ddf9368d76e3eb0b70b862977542e266869f91f0ac2738c8042b2ff0591"},
	{"events/create/INSERT_event_description.sql", 103, "925802813df4777a633daf3fd6ae0331577416eca0292bcc02eaff993488205d"},
	{"events/images/confirm/INSERT_event_image.sql", 74, "7cfedca2b318385f7747228674b9828c4dbc985d7db445df7e0dea4c410245cd"},
	{"events/images/confirm/INSERT_image.sql", 118, "2e9b6bd067ed1648317349723f46820c34a780dd477737d38fb3319d8dc00e28"},
	{"events/read/SELECT_events.sql", 753, "b4a60628b06a16a4e5027a5e056aaeebb447c985a9ef344fda567d58fee53444"},
	{"me/clubs/list/SELECT_student_clubs.sql", 385, "bc6b4333e217b235d570097fed286bf7052256eefdcc65fcbc545602fc57d18f"},
	{"me/events/list/SELECT_student_events.sql", 835, "dbc8ecb401936483e66a7a3e40b31615bacfe68e5cab21333c95f96678599943"},
	{"me/get/SELECT_student_by_sub.sql", 189, "abe30d5eded05b5bea2162a1e93fcee66b20142bb47e5e0fac8b5833a96ddc47"},
	{"students/ensure/EXISTS_student_by_sub.sql", 278, "1c36fc43596da415f30e3e15d88a2659fb02a22761209e88e0b4adc8e82a3a96"},
	{"students/ensure/UPSERT_student.sql", 146, "b281775ddf5418ee7d6ed0b744b13740ba917b5072e84f6b01d95ad1462be9fb"},
	{"students/ensure/UPSERT_student_sub_only.sql", 239, "1b5ec7e066fc3461115d756d2cf8d1c1cfe73e26b7a038574e7ff175b4b04d06"},
}
